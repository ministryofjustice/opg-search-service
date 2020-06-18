package elasticsearch

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Client struct {
	logger *log.Logger
}

type Indexable interface {
	GetIndexName() string
	GetJson() string
}

func NewClient(logger *log.Logger) (*Client, error) {
	return &Client{logger: logger}, nil
}

func (c Client) Index(i Indexable) error {
	// Basic information for the Amazon Elasticsearch Service domain
	domain := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT") // e.g. https://my-domain.region.es.amazonaws.com
	endpoint := domain + "/" + i.GetIndexName() + "/" + "_doc"

	var region string
	var ok bool
	if region, ok = os.LookupEnv("AWS_REGION"); !ok {
		region = "eu-west-1"
	}
	service := "es"

	body := strings.NewReader(i.GetJson())

	// Get credentials from environment variables and create the AWS Signature Version 4 signer
	cred := credentials.NewEnvCredentials()
	signer := v4.NewSigner(cred)

	// An HTTP client for sending the request
	client := &http.Client{}

	// Form the HTTP request
	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return err
	}

	// You can probably infer Content-Type programmatically, but here, we just say that it's JSON
	req.Header.Add("Content-Type", "application/json")

	// Sign the request, send it, and print the response
	signer.Sign(req, body, service, region, time.Now())

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return errors.New(string(bodyBytes))
	}

	c.logger.Println(resp.Status + "\n")
	return nil
}
