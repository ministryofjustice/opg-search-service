package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/sirupsen/logrus"
)

const maxPayloadSize = 10485760 // bytes
const backoff = 6 * time.Second
const maxRetries = 10

var ErrAliasMissing = errors.New("alias is missing")
var ErrOpTooLarge = errors.New("BulkOp exceeds maximum payload size")
var errTooManyRequests = errors.New("too many requests")

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
			Source json.RawMessage `json:"_source"`
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
	client := &Client{
		httpClient: httpClient,
		logger:     logger,
		domain:     os.Getenv("AWS_ELASTICSEARCH_ENDPOINT"),
		region:     os.Getenv("AWS_REGION"),
		service:    "es",
		signer:     v4.NewSigner(credentials.NewEnvCredentials()),
	}

	if client.region == "" {
		client.region = "eu-west-1"
	}

	return client, nil
}

func (c *Client) doRequest(method, endpoint string, body io.ReadSeeker, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.domain+"/"+endpoint, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	_, _ = c.signer.Sign(req, body, c.service, c.region, time.Now())

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

func (op *BulkOp) Index(id int64, v interface{}) error {
	if err := op.enc.Encode(bulkOp{Index: indexOp{ID: strconv.Itoa(int(id))}}); err != nil {
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

func (c *Client) DoBulk(op *BulkOp) (BulkResult, error) {
	retries := 0

	for {
		res, err := c.doBulkOp(op)
		if err == errTooManyRequests && retries < maxRetries {
			retries++
			time.Sleep(time.Duration(retries) * backoff)
			continue
		}

		return res, err
	}
}

func (c *Client) doBulkOp(op *BulkOp) (BulkResult, error) {
	body := bytes.NewReader(op.buf.Bytes())

	endpoint := fmt.Sprintf("%s/_bulk", op.index)
	resp, err := c.doRequest(http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		c.logger.Error(err.Error())

		return BulkResult{}, fmt.Errorf("unable to process index request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return BulkResult{}, errTooManyRequests
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
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
func (c *Client) Search(indices []string, requestBody map[string]interface{}) (*SearchResult, error) {
	endpoint := strings.Join(indices, ",") + "/_search"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return nil, err
	}
	body := bytes.NewReader(buf.Bytes())

	resp, err := c.doRequest(http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
		hits[i] = hit.Source
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

func (c *Client) CreateIndex(name string, config []byte, force bool) error {
	exists, err := c.IndexExists(name)
	if err != nil {
		return err
	}

	if exists {
		if !force {
			c.logger.Printf("index '%s' already exists", name)
			return nil
		}

		c.logger.Printf("changes are forced, deleting old index '%s'", name)

		if err := c.DeleteIndex(name); err != nil {
			return err
		}

		c.logger.Printf("index '%s' deleted", name)
	}

	if err := c.createIndex(name, config); err != nil {
		return err
	}

	c.logger.Printf("index '%s' created", name)
	return nil
}

func (c *Client) IndexExists(name string) (bool, error) {
	c.logger.Printf("Checking index '%s' exists", name)

	resp, err := c.doRequest(http.MethodHead, name, nil, "")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, fmt.Errorf("index check failed with status code %d", resp.StatusCode)
}

func (c *Client) createIndex(name string, config []byte) error {
	c.logger.Printf("Creating index '%s'", name)

	resp, err := c.doRequest(http.MethodPut, name, bytes.NewReader(config), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf(`index creation failed with status code %d and response: "%s"`, resp.StatusCode, string(data)))
	}

	return nil
}

func (c *Client) DeleteIndex(name string) error {
	c.logger.Printf("Deleting index '%s'", name)

	resp, err := c.doRequest(http.MethodDelete, name, nil, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) ResolveAlias(name string) (string, error) {
	resp, err := c.doRequest(http.MethodGet, "/_alias/"+name, nil, "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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

func (c *Client) CreateAlias(alias, index string) error {
	resp, err := c.doRequest(http.MethodPut, fmt.Sprintf("%s/_alias/%s", index, alias), nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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

func (c *Client) UpdateAlias(alias, oldIndex, newIndex string) error {
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

	resp, err := c.doRequest(http.MethodPost, "_aliases", bytes.NewReader(request), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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

func (c *Client) Indices(term string) ([]string, error) {
	resp, err := c.doRequest(http.MethodGet, term, nil, "")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (c *Client) Delete(indices []string, requestBody map[string]interface{}) (*DeleteResult, error) {
	endpoint := strings.Join(indices, ",") + "/_delete_by_query"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return nil, err
	}
	body := bytes.NewReader(buf.Bytes())

	c.logger.Info(buf.String())

	resp, err := c.doRequest(http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
