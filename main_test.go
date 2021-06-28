package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"opg-search-service/response"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite struct {
	suite.Suite
	testPeople []person.Person
	esClient   elasticsearch.ClientInterface
	authHeader string
}

func (suite *EndToEndTestSuite) SetupSuite() {
	os.Setenv("JWT_SECRET", "MyTestSecret")
	os.Setenv("USER_HASH_SALT", "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0")
	os.Setenv("ENVIRONMENT", "local")

	logger, _ := test.NewNullLogger()
	httpClient := &http.Client{}
	suite.esClient, _ = elasticsearch.NewClient(httpClient, logger)

	suite.authHeader = "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6OTk5OTk5OTk5OSwic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.8HtN6aTAnE2YFI9rJD8drzqgrXPkyUbwRRJymcPSmHk"

	// define fixtures
	var ids []int64
	for i := 0; i < 2; i++ {
		ids = append(ids, int64(i))
		suite.testPeople = append(suite.testPeople, person.Person{
			ID:         &ids[i],
			Firstname:  fmt.Sprintf("John%d", i),
			Surname:    fmt.Sprintf("Doe%d", i),
			Persontype: fmt.Sprintf("Type%d", i%2),
		})
	}

	// wait for ES service to stand up
	time.Sleep(time.Second * 10)

	// start the app
	go main()

	// delete all indices
	req, _ := http.NewRequest(http.MethodDelete, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/_all", nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	suite.NotNil(resp)
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	exists, err := suite.esClient.IndexExists(person.Person{})
	suite.False(exists, "Person index should not exist at this point")
	suite.Nil(err)

	// wait up to 5 seconds for the app to start
	retries := 5
	for i := 1; i <= retries; i++ {
		conn, err := net.DialTimeout("tcp", "localhost:8000", time.Second)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		exists, err = suite.esClient.IndexExists(person.Person{})
		suite.True(exists, "Person index should exist at this point")
		suite.Nil(err)

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
		Persons: suite.testPeople,
	}

	jsonBody, _ := json.Marshal(iReq)
	reqBody := bytes.NewReader(jsonBody)
	req, _ := http.NewRequest(http.MethodPost, suite.GetUrl("/persons"), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.authHeader)

	resp, err := client.Do(req)
	if err != nil {
		suite.Fail("Error indexing person", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusAccepted, resp.StatusCode)

	data, _ := ioutil.ReadAll(resp.Body)

	suite.Equal(`{"results":[{"id":0,"statusCode":201,"message":""},{"id":1,"statusCode":201,"message":""}]}`, string(data))

	hit, _ := json.Marshal(suite.testPeople[1])

	expectedSearchResp, _ := json.Marshal(response.SearchResponse{
		Results: []json.RawMessage{hit},
		Aggregations: map[string]map[string]int{
			"personType": {
				"Type1": 1,
			},
		},
		Total: response.Total{
			Count: 1,
			Exact: true,
		},
	})

	reqBody = bytes.NewReader([]byte(`{"term":"` + suite.testPeople[1].Surname + `"}`))
	req, _ = http.NewRequest(http.MethodPost, suite.GetUrl("/persons/search"), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.authHeader)

	var respBody string

	// wait up to 2s for the indexed record to become searchable
	for i := 0; i < 20; i++ {
		resp, err := client.Do(req)
		if err != nil {
			suite.Fail("Error searching for a person", err)
		}

		suite.Equal(http.StatusOK, resp.StatusCode)

		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		respBody = buf.String()

		if string(expectedSearchResp) == respBody {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	suite.Equal(string(expectedSearchResp), respBody, "Unexpected search result")
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
