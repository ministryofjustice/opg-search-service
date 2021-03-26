package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"opg-search-service/response"
	"os"
	"testing"
	"time"
)

type EndToEndTestSuite struct {
	suite.Suite
	testPerson *person.Person
	esClient   *elasticsearch.Client
}

func (suite *EndToEndTestSuite) SetupSuite() {
	logBuf := new(bytes.Buffer)
	logger := log.New(logBuf, "opg-file-service ", log.LstdFlags)
	suite.esClient, _ = elasticsearch.NewClient(logger)

	// define fixtures
	suite.testPerson = &person.Person{
		Id:        3,
		FirstName: "John",
		LastName:  "Doe",
	}

	// start the app
	go main()

	log.Printf("Waiting up to 30 seconds for the search service and ElasticSearch servers to start")
	ssUp := false
	esUp := false
	retries := 30
	for i := 1; i <= retries; i++ {
		if !ssUp {
			conn, err := net.DialTimeout("tcp", "localhost:8000", time.Second)
			if err == nil {
				ssUp = true
				conn.Close()
			}
		}
		if !esUp {
			conn, err := net.DialTimeout("tcp", os.Getenv("AWS_ELASTICSEARCH_ENDPOINT"), time.Second)
			if err == nil {
				esUp = true
				conn.Close()
			}
		}

		if ssUp && esUp {
			return
		}

		log.Printf("Search service: '%t', ElasticSearch: '%t' after %d retries", ssUp, esUp, i)
		time.Sleep(time.Second)
	}

	suite.Fail(fmt.Sprintf("Unable to start search service and ElasticSearch after %d attempts", retries))
}

func (suite *EndToEndTestSuite) GetUrl(path string) string {
	return "http://localhost:8000" + os.Getenv("PATH_PREFIX") + path
}

func (suite *EndToEndTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.GetUrl("/health-check"))
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *EndToEndTestSuite) TestIndexPerson() {
	client := new(http.Client)

	iReq := person.IndexRequest{
		Persons: []person.Person{
			*suite.testPerson,
		},
	}

	jsonBody, _ := json.Marshal(iReq)
	reqBody := bytes.NewReader(jsonBody)
	req, _ := http.NewRequest(http.MethodPost, suite.GetUrl("/persons"), reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		suite.Fail("Error indexing a person", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusAccepted, resp.StatusCode)

	var iResp response.IndexResponse

	err = json.NewDecoder(resp.Body).Decode(&iResp)
	if err != nil {
		suite.Fail("Unable to decode JSON response", resp.Body)
	}

	expectedResp := response.IndexResponse{
		Results: []elasticsearch.IndexResult{
			{
				Id:         suite.testPerson.Id,
				StatusCode: 201,
				Message:    "Index created",
			},
		},
	}

	suite.Equal(expectedResp, iResp, "Unexpected index result")
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
