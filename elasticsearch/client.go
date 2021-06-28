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
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/sirupsen/logrus"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientInterface interface {
	DoBulk(op *BulkOp) *BulkResult
	Search(requestBody map[string]interface{}, dataType Indexable) (*SearchResult, error)
	CreateIndex(i Indexable) (bool, error)
	DeleteIndex(i Indexable) error
	IndexExists(i Indexable) (bool, error)
}

type Indexable interface {
	Id() int64
	IndexName() string
	Json() string
	IndexConfig() map[string]interface{}
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
	req, err := http.NewRequest(method, endpoint, body)
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
	enc   *json.Encoder
}

func NewBulkOp(index string) *BulkOp {
	op := &BulkOp{}
	op.index = index
	op.enc = json.NewEncoder(&op.buf)
	return op
}

func (op *BulkOp) Index(id int64, v interface{}) error {
	if err := op.enc.Encode(bulkOp{Index: indexOp{ID: strconv.Itoa(int(id))}}); err != nil {
		return err
	}

	return op.enc.Encode(v)
}

func (c *Client) DoBulk(op *BulkOp) *BulkResult {
	body := bytes.NewReader(op.buf.Bytes())

	endpoint := fmt.Sprintf("%s/%s/_bulk", c.domain, op.index)
	resp, err := c.doRequest(http.MethodPost, endpoint, body, "application/json")
	if err != nil {
		c.logger.Error(err.Error())

		return &BulkResult{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to process document index request",
		}
	}
	defer resp.Body.Close()

	res := &BulkResult{StatusCode: resp.StatusCode}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		c.logger.Error(string(bodyBytes))

		res.Message = string(bodyBytes)
		return res
	}

	var v bulkResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		res.Message = err.Error()
		return res
	}

	if v.Errors {
		for _, d := range v.Items {
			res.Results = append(res.Results, BulkResultItem{ID: d.Index.ID, StatusCode: d.Index.Status})
		}
	}

	return res
}

// returns an array of JSON encoded results
func (c *Client) Search(requestBody map[string]interface{}, dataType Indexable) (*SearchResult, error) {
	endpoint := c.domain + "/" + dataType.IndexName() + "/_search"

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

func (c *Client) IndexExists(i Indexable) (bool, error) {
	c.logger.Printf("Checking index '%s' exists", i.IndexName())

	endpoint := c.domain + "/" + i.IndexName()

	body := bytes.NewReader([]byte(""))
	resp, err := c.doRequest(http.MethodHead, endpoint, body, "")
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	}

	return false, errors.New(fmt.Sprintf(`index check failed with status code %d`, resp.StatusCode))
}

func (c *Client) CreateIndex(i Indexable) (bool, error) {
	c.logger.Printf("Creating index '%s' for %T", i.IndexName(), i)

	endpoint := c.domain + "/" + i.IndexName()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(i.IndexConfig()); err != nil {
		return false, err
	}
	body := bytes.NewReader(buf.Bytes())

	resp, err := c.doRequest(http.MethodPut, endpoint, body, "application/json")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf.Reset()
		_, _ = buf.ReadFrom(resp.Body)
		return false, errors.New(fmt.Sprintf(`index creation failed with status code %d and response: "%s"`, resp.StatusCode, buf.String()))
	}

	return true, nil
}

func (c *Client) DeleteIndex(i Indexable) error {
	c.logger.Printf("Deleting index '%s' for %T", i.IndexName(), i)

	endpoint := c.domain + "/" + i.IndexName()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(i.IndexConfig()); err != nil {
		return err
	}
	body := bytes.NewReader(buf.Bytes())

	resp, err := c.doRequest(http.MethodDelete, endpoint, body, "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
