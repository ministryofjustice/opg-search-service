package elasticsearch

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientInterface interface {
	Index(i Indexable) *IndexResult
}

type Indexable interface {
	Id() int64
	IndexName() string
	Json() string
}

type Client struct {
	httpClient HTTPClient
	logger     *log.Logger
}

func NewClient(httpClient HTTPClient, logger *log.Logger) (ClientInterface, error) {
	return &Client{
		httpClient: httpClient,
		logger:     logger,
	}, nil
}

func (c Client) Index(i Indexable) *IndexResult {
	c.logger.Printf("Indexing %T ID %d", i, i.Id())

	// Basic information for the Amazon Elasticsearch Service domain
	domain := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT") // e.g. https://my-domain.region.es.amazonaws.com
	endpoint := domain + "/" + i.IndexName() + "/_doc/" + strconv.FormatInt(i.Id(), 10)

	var region string
	var ok bool
	if region, ok = os.LookupEnv("AWS_REGION"); !ok {
		region = "eu-west-1"
	}
	service := "es"

	body := strings.NewReader(i.Json())

	// Get credentials from environment variables and create the AWS Signature Version 4 signer
	cred := credentials.NewEnvCredentials()
	signer := v4.NewSigner(cred)

	iRes := IndexResult{Id: i.Id()}

	// Form the HTTP request
	req, err := http.NewRequest(http.MethodPut, endpoint, body)
	if err != nil {
		c.logger.Println(err.Error())
		iRes.StatusCode = http.StatusInternalServerError
		iRes.Message = "Unable to create index request"
		return &iRes
	}

	// You can probably infer Content-Type programmatically, but here, we just say that it's JSON
	req.Header.Add("Content-Type", "application/json")

	// Sign the request, send it, and print the response
	_, _ = signer.Sign(req, body, service, region, time.Now())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Println(err.Error())
		iRes.StatusCode = http.StatusInternalServerError
		iRes.Message = "Unable to process index request"
		return &iRes
	}

	iRes.StatusCode = resp.StatusCode

	switch iRes.StatusCode {
	case http.StatusOK:
		iRes.Message = "Index updated"
	case http.StatusCreated:
		iRes.Message = "Index created"
	default:
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		iRes.Message = string(bodyBytes)
	}

	return &iRes
}
