package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"opg-search-service/response"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite struct {
	suite.Suite
	testPerson *person.Person
	esClient   elasticsearch.ClientInterface
	authHeader string
}

func (suite *EndToEndTestSuite) SetupSuite() {
	os.Setenv("JWT_SECRET", "MyTestSecret")
	os.Setenv("USER_HASH_SALT", "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0")

	logBuf := new(bytes.Buffer)
	logger := log.New(logBuf, "opg-file-service ", log.LstdFlags)
	httpClient := &http.Client{}
	suite.esClient, _ = elasticsearch.NewClient(httpClient, logger)

	suite.authHeader = "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6OTk5OTk5OTk5OSwic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.8HtN6aTAnE2YFI9rJD8drzqgrXPkyUbwRRJymcPSmHk"

	// define fixtures
	id := int64(3)
	suite.testPerson = &person.Person{
		UID:           "3",
		Normalizeduid: &id,
		Firstname:     "John",
		Surname:       "Doe",
	}

	// wait for ES service to stand up
	time.Sleep(time.Second * 10)

	// start the app
	go main()

	// delete all indices
	req, _ := http.NewRequest(http.MethodDelete, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/_all", nil)
	resp, err := httpClient.Do(req)
	suite.NotNil(resp)
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	// create indices
	ok, err := suite.esClient.CreateIndex(person.Person{})
	suite.True(ok, "Could not create Person index")
	suite.Nil(err)

	// wait up to 5 seconds for the app to start
	retries := 5
	for i := 1; i <= retries; i++ {
		conn, err := net.DialTimeout("tcp", "localhost:8000", time.Second)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		conn.Close()
		return
	}

	suite.Fail(fmt.Sprintf("Unable to start search service server after %d attempts", retries))
}

func (suite *EndToEndTestSuite) GetUrl(path string) string {
	return "http://localhost:8000" + os.Getenv("PATH_PREFIX") + path
}

func (suite *EndToEndTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.GetUrl("/health-check"))
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *EndToEndTestSuite) TestIndexAndSearchPerson() {
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
	req.Header.Set("Authorization", suite.authHeader)

	resp, err := client.Do(req)
	if err != nil {
		suite.Fail("Error indexing a person", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusAccepted, resp.StatusCode)

	var iResp response.IndexResponse

	err = json.NewDecoder(resp.Body).Decode(&iResp)
	if err != nil {
		suite.Fail("Unable to decode JSON index response", resp.Body)
	}

	expectedResp := response.IndexResponse{
		Results: []elasticsearch.IndexResult{
			{
				Id:         3,
				StatusCode: 201,
				Message:    "Index created",
			},
		},
	}

	suite.Equal(expectedResp, iResp, "Unexpected index result")

	time.Sleep(time.Second * 2)

	sReq := person.SearchRequest{
		Term: suite.testPerson.Surname,
	}

	jsonBody, _ = json.Marshal(sReq)
	reqBody = bytes.NewReader(jsonBody)
	req, _ = http.NewRequest(http.MethodPost, suite.GetUrl("/persons/search"), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.authHeader)

	resp, err = client.Do(req)
	if err != nil {
		suite.Fail("Error searching for a person", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)

	expectedSearchResp, _ := json.Marshal(response.SearchResponse{
		Results: []elasticsearch.Indexable{
			suite.testPerson,
		},
	})

	suite.Equal(expectedSearchResp, buf.Bytes(), "Unexpected search result")
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
