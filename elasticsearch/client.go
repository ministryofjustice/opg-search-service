package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientInterface interface {
	Index(i Indexable) *IndexResult
	Search(requestBody map[string]interface{}, dataType Indexable) ([][]byte, error)
	CreateIndex(i Indexable) (bool, error)
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
	logger     *log.Logger
	domain     string
	region     string
	service    string
	signer     *v4.Signer
}

type elasticSearchResponse struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"eq"`
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

func NewClient(httpClient HTTPClient, logger *log.Logger) (ClientInterface, error) {
	client := Client{
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

	return &client, nil
}

func (c Client) Index(i Indexable) *IndexResult {
	c.logger.Printf("Indexing %T ID %d", i, i.Id())

	endpoint := c.domain + "/" + i.IndexName() + "/_doc/" + strconv.FormatInt(i.Id(), 10)

	body := strings.NewReader(i.Json())

	iRes := IndexResult{Id: i.Id()}

	// Form the HTTP request
	req, err := http.NewRequest(http.MethodPut, endpoint, body)
	if err != nil {
		c.logger.Println(err.Error())
		iRes.StatusCode = http.StatusInternalServerError
		iRes.Message = "Unable to create document index request"
		return &iRes
	}

	// You can probably infer Content-Type programmatically, but here, we just say that it's JSON
	req.Header.Add("Content-Type", "application/json")

	// Sign the request, send it, and print the response
	_, _ = c.signer.Sign(req, body, c.service, c.region, time.Now())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Println(err.Error())
		iRes.StatusCode = http.StatusInternalServerError
		iRes.Message = "Unable to process document index request"
		return &iRes
	}

	iRes.StatusCode = resp.StatusCode

	switch iRes.StatusCode {
	case http.StatusOK:
		iRes.Message = "Document updated"
	case http.StatusCreated:
		iRes.Message = "Document created"
	default:
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		iRes.Message = string(bodyBytes)
	}

	return &iRes
}

// returns an array of JSON encoded results
func (c Client) Search(requestBody map[string]interface{}, dataType Indexable) ([][]byte, error) {
	endpoint := c.domain + "/" + dataType.IndexName() + "/_search"

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(requestBody); err != nil {
		return nil, err
	}
	body := bytes.NewReader(buf.Bytes())

	// Form the HTTP request
	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}

	// You can probably infer Content-Type programmatically, but here, we just say that it's JSON
	req.Header.Add("Content-Type", "application/json")

	// Sign the request, send it, and print the response
	_, _ = c.signer.Sign(req, body, c.service, c.region, time.Now())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		buf.Reset()
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf(`search request failed with status code %d and response: "%s"`, resp.StatusCode, buf.String())
	}

	var esResponse elasticSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&esResponse); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %w", err)
	}

	var results [][]byte
	for _, hit := range esResponse.Hits.Hits {
		results = append(results, bytes.TrimSpace(hit.Source))
	}

	return results, nil
}

func (c Client) IndexExists(i Indexable) (bool, error) {
	c.logger.Printf("Checking index '%s' exists", i.IndexName())

	endpoint := c.domain + "/" + i.IndexName()

	// Form the HTTP request
	body := bytes.NewReader([]byte(""))
	req, err := http.NewRequest(http.MethodHead, endpoint, body)
	if err != nil {
		return false, err
	}

	// Sign the request, send it, and print the response
	_, _ = c.signer.Sign(req, body, c.service, c.region, time.Now())

	resp, err := c.httpClient.Do(req)
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

func (c Client) CreateIndex(i Indexable) (bool, error) {
	c.logger.Printf("Creating index '%s' for %T", i.IndexName(), i)

	endpoint := c.domain + "/" + i.IndexName()

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(i.IndexConfig()); err != nil {
		return false, err
	}
	body := bytes.NewReader(buf.Bytes())

	// Form the HTTP request
	req, err := http.NewRequest(http.MethodPut, endpoint, body)
	if err != nil {
		return false, err
	}

	// You can probably infer Content-Type programmatically, but here, we just say that it's JSON
	req.Header.Add("Content-Type", "application/json")

	// Sign the request, send it, and print the response
	_, _ = c.signer.Sign(req, body, c.service, c.region, time.Now())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		buf.Reset()
		_, _ = buf.ReadFrom(resp.Body)
		return false, errors.New(fmt.Sprintf(`index creation failed with status code %d and response: "%s"`, resp.StatusCode, buf.String()))
	}

	return true, nil
}
