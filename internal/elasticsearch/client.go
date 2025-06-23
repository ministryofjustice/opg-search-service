package elasticsearch

import (
	"bytes"
	"context"
	"crypto/sha256"

	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/sirupsen/logrus"
)

const maxPayloadSize = 10485760 // bytes
const backoff = 6 * time.Second
const maxRetries = 10

var ErrAliasMissing = errors.New("alias is missing")
var ErrOpTooLarge = errors.New("BulkOp exceeds maximum payload size")
var errTooManyRequests = errors.New("too many requests")

// to replace "poadraftapplication_d339d717c2035a54" style index aliases
// with the start of the alias, e.g. "poadraftapplication"
var indexAliasCleaner = regexp.MustCompile("_(.*)$")

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
	logger     *logrus.Logger
	domain     string
	region     string
	service    string
	signer     *v4.Signer
}

type elasticSearchResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			Index  string                 `json:"_index"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]struct {
		Buckets []struct {
			Key      string `json:"key"`
			DocCount int    `json:"doc_count"`
		} `json:"buckets"`
	} `json:"aggregations"`
}

type SearchResult struct {
	Hits         []json.RawMessage
	Aggregations map[string]map[string]int
	Total        int
	TotalExact   bool
}

type DeleteResult struct {
	Total int
}

func NewClient(httpClient HTTPClient, logger *logrus.Logger) (*Client, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-west-1"
	}

	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))

	client := &Client{
		httpClient: httpClient,
		logger:     logger,
		domain:     os.Getenv("AWS_ELASTICSEARCH_ENDPOINT"),
		region:     region,
		service:    os.Getenv("AWS_SEARCH_PROVIDER"),
		signer:     v4.NewSigner(sess.Config.Credentials),
	}

	if client.service == "" {
		client.service = "es"
	}

	return client, nil
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body io.ReadSeeker, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.domain+"/"+endpoint, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(body)
		if err != nil {
			return nil, err
		}

		_, err = body.Seek(0,io.SeekStart)
		if err != nil {
			return nil, err
		}

		h := sha256.New()
		h.Write(buf.Bytes())
		bs := hex.EncodeToString(h.Sum(nil))
		req.Header.Add("X-Amz-Content-Sha256", bs)
		} else {
			req.Header.Add("X-Amz-Content-Sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	_, err = c.signer.Sign(req, body, c.service, c.region, time.Now())
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Status int    `json:"status"`
			Error  *struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

type bulkOp struct {
	Index indexOp `json:"index"`
}

type indexOp struct {
	ID string `json:"_id"`
}

type BulkOp struct {
	index string
	buf   bytes.Buffer
	tmp   bytes.Buffer
	enc   *json.Encoder
}

func NewBulkOp(index string) *BulkOp {
	op := &BulkOp{}
	op.index = index
	op.enc = json.NewEncoder(&op.tmp)
	return op
}

func (op *BulkOp) Index(id string, v interface{}) error {
	if err := op.enc.Encode(bulkOp{Index: indexOp{ID: id}}); err != nil {
		return err
	}

	if err := op.enc.Encode(v); err != nil {
		return err
	}

	if op.tmp.Len()+op.buf.Len() > maxPayloadSize {
		return ErrOpTooLarge
	}

	_, err := op.tmp.WriteTo(&op.buf)
	return err
}

func (op *BulkOp) Empty() bool {
	return op.buf.Len() == 0
}

func (op *BulkOp) Reset() {
	op.buf.Reset()
	op.tmp.Reset()
}

type BulkResult struct {
	Successful int
	Failed     int
	Error      string
}

func (c *Client) DoBulk(ctx context.Context, op *BulkOp) (BulkResult, error) {
	retries := 0

	for {
		res, err := c.doBulkOp(ctx, op)
		if err == errTooManyRequests && retries < maxRetries {
			retries++
			time.Sleep(time.Duration(retries) * backoff)
			continue
		}

		return res, err
	}
}

func (c *Client) doBulkOp(ctx context.Context, op *BulkOp) (BulkResult, error) {
	body := bytes.NewReader(op.buf.Bytes())

	endpoint := fmt.Sprintf("%s/_bulk", op.index)
	resp, err := c.doRequest(ctx, http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		c.logger.Error(err.Error())

		return BulkResult{}, fmt.Errorf("unable to process index request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	if resp.StatusCode == http.StatusTooManyRequests {
		return BulkResult{}, errTooManyRequests
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Error(string(bodyBytes))

		return BulkResult{}, fmt.Errorf("elasticsearch failed: %s", string(bodyBytes))
	}

	var v bulkResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return BulkResult{}, fmt.Errorf("deserializing response: %s", err)
	}

	var result BulkResult
	for _, d := range v.Items {
		if d.Index.Status == http.StatusOK || d.Index.Status == http.StatusCreated {
			result.Successful += 1
		} else {
			result.Failed += 1
			if d.Index.Error != nil && result.Error == "" {
				result.Error = fmt.Sprintf("%s: %s", d.Index.Error.Type, d.Index.Error.Reason)
			}
		}
	}

	return result, nil
}

// returns an array of JSON encoded results
func (c *Client) Search(ctx context.Context, indices []string, requestBody map[string]interface{}) (*SearchResult, error) {
	endpoint := strings.Join(indices, ",") + "/_search"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return nil, err
	}
	body := bytes.NewReader(buf.Bytes())

	resp, err := c.doRequest(ctx, http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	if resp.StatusCode != http.StatusOK {
		buf.Reset()
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf(`search request failed with status code %d and response: "%s"`, resp.StatusCode, buf.String())
	}

	var esResponse elasticSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResponse); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %w", err)
	}

	hits := make([]json.RawMessage, len(esResponse.Hits.Hits))
	for i, hit := range esResponse.Hits.Hits {
		hit.Source["_index"] = indexAliasCleaner.ReplaceAllString(hit.Index, "")

		result, err := json.Marshal(hit.Source)
		if err != nil {
			return nil, fmt.Errorf("unable to add _index to JSON: %w", err)
		}

		hits[i] = result
	}

	aggregations := map[string]map[string]int{}
	for field, v := range esResponse.Aggregations {
		for _, bucket := range v.Buckets {
			if m, ok := aggregations[field]; ok {
				m[bucket.Key] = bucket.DocCount
			} else {
				aggregations[field] = map[string]int{bucket.Key: bucket.DocCount}
			}
		}
	}

	return &SearchResult{
		Hits:         hits,
		Aggregations: aggregations,
		Total:        esResponse.Hits.Total.Value,
		TotalExact:   esResponse.Hits.Total.Relation == "eq",
	}, nil
}

func (c *Client) CreateIndex(ctx context.Context, name string, config []byte, force bool) error {
	c.logger.Printf("Checking index '%s' exists", name)
	exists, err := c.IndexExists(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		if !force {
			c.logger.Printf("index '%s' already exists", name)
			return nil
		}

		c.logger.Printf("changes are forced, deleting old index '%s'", name)

		if err := c.DeleteIndex(ctx, name); err != nil {
			return err
		}

		c.logger.Printf("index '%s' deleted", name)
		time.Sleep(20 * time.Second)
	}

	if err := c.createIndex(ctx, name, config); err != nil {
		return err
	}

	c.logger.Printf("index '%s' created", name)
	return nil
}

func (c *Client) IndexExists(ctx context.Context, name string) (bool, error) {
	resp, err := c.doRequest(ctx, http.MethodHead, name, nil, "")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, fmt.Errorf("index check failed with status code %d", resp.StatusCode)
}

func (c *Client) createIndex(ctx context.Context, name string, config []byte) error {
	c.logger.Printf("Creating index '%s'", name)

	resp, err := c.doRequest(ctx, http.MethodPut, name, bytes.NewReader(config), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(`index creation failed with status code %d and response: "%s"`, resp.StatusCode, string(data))
	}

	return nil
}

func (c *Client) DeleteIndex(ctx context.Context, name string) error {
	c.logger.Printf("Deleting index '%s'", name)

	resp, err := c.doRequest(ctx, http.MethodDelete, name, nil, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	return nil
}

func (c *Client) ResolveAlias(ctx context.Context, name string) (string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "_alias/"+name, nil, "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrAliasMissing
	}

	var v map[string]struct{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}

	for k := range v {
		return k, nil
	}

	return "", ErrAliasMissing
}

func (c *Client) CreateAlias(ctx context.Context, alias, index string) error {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("%s/_alias/%s", index, alias), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	var v struct {
		Acknowledged bool `json:"acknowledged"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return err
	}

	if !v.Acknowledged {
		return fmt.Errorf("problem creating alias '%s' for index '%s'", alias, index)
	}

	c.logger.Printf("alias '%s' for index '%s' created", alias, index)
	return nil
}

type aliasRequest struct {
	Actions []map[string]aliasRequestAction `json:"actions"`
}

type aliasRequestAction struct {
	Index string `json:"index"`
	Alias string `json:"alias"`
}

func (c *Client) UpdateAlias(ctx context.Context, alias, oldIndex, newIndex string) error {
	c.logger.Printf("Updating alias '%s' from index '%s' to '%s'", alias, oldIndex, newIndex)

	request, err := json.Marshal(aliasRequest{
		Actions: []map[string]aliasRequestAction{
			{"remove": {
				Alias: alias,
				Index: oldIndex,
			}},
			{"add": {
				Alias: alias,
				Index: newIndex,
			}},
		},
	})
	if err != nil {
		return err
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "_aliases", bytes.NewReader(request), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	var v struct {
		Acknowledged bool `json:"acknowledged"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return err
	}

	if !v.Acknowledged {
		return fmt.Errorf("problem updating alias '%s' for index '%s'", alias, newIndex)
	}

	return nil
}

func (c *Client) Indices(ctx context.Context, term string) ([]string, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, term, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	var v map[string]struct{}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	var ks []string
	for k := range v {
		ks = append(ks, k)
	}

	return ks, nil
}

func (c *Client) Delete(ctx context.Context, indices []string, requestBody map[string]interface{}) (*DeleteResult, error) {
	endpoint := strings.Join(indices, ",") + "/_delete_by_query?conflicts=proceed"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return nil, err
	}
	body := bytes.NewReader(buf.Bytes())

	resp, err := c.doRequest(ctx, http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	if resp.StatusCode != http.StatusOK {
		buf.Reset()
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf(`delete request failed with status code %d and response: "%s"`, resp.StatusCode, buf.String())
	}

	var esResponse struct {
		Total int `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&esResponse); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %w", err)
	}

	return &DeleteResult{
		Total: esResponse.Total,
	}, nil
}
