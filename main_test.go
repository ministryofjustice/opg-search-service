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

	suite.testPeople = []person.Person{
		{
			ID:         id(0),
			Firstname:  "John0",
			Surname:    "Doe0",
			Persontype: "Type0",
			Dob:        "01/02/1990",
			Addresses: []person.PersonAddress{{
				Postcode: "NG1 2CD",
			}},
		},
		{
			ID:         id(1),
			Firstname:  "John1",
			Surname:    "Doe1",
			Persontype: "Type1",
			Dob:        "20/03/1987",
			Addresses: []person.PersonAddress{{
				Postcode: "NG1 1AB",
			}},
		},
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

func (suite *EndToEndTestSuite) TestHealthCheck() {
	resp, err := http.Get("http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check")
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *EndToEndTestSuite) TestIndexAndSearchPerson() {
	resp, err := doRequest(suite.authHeader, "/persons", person.IndexRequest{Persons: suite.testPeople})
	if err != nil {
		suite.Fail("Error indexing person", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusAccepted, resp.StatusCode)

	data, _ := ioutil.ReadAll(resp.Body)

	suite.Equal(`{"successful":2,"failed":0}`, string(data))

	testCases := []struct {
		scenario         string
		term             string
		expectedResponse func() response.SearchResponse
	}{
		{
			scenario: "search by surname",
			term:     suite.testPeople[1].Surname,
			expectedResponse: func() response.SearchResponse {
				hit, _ := json.Marshal(suite.testPeople[1])

				return response.SearchResponse{
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
				}
			},
		},
		{
			scenario: "search by dob",
			term:     "01/02/1990",
			expectedResponse: func() response.SearchResponse {
				hit, _ := json.Marshal(suite.testPeople[0])

				return response.SearchResponse{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type0": 1,
						},
					},
					Total: response.Total{
						Count: 1,
						Exact: true,
					},
				}
			},
		},
		{
			scenario: "search by postcode",
			term:     "NG1 2CD",
			expectedResponse: func() response.SearchResponse {
				hit, _ := json.Marshal(suite.testPeople[0])

				return response.SearchResponse{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type0": 1,
						},
					},
					Total: response.Total{
						Count: 1,
						Exact: true,
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.scenario, func() {
			var respBody []byte

			expectedResponse, _ := json.Marshal(tc.expectedResponse())

			// wait up to 2s for the indexed record to become searchable
			for i := 0; i < 20; i++ {
				resp, err := doRequest(suite.authHeader, "/persons/search", map[string]string{"term": tc.term})
				if err != nil {
					suite.Fail("Error searching for a person", err)
				}

				suite.Equal(http.StatusOK, resp.StatusCode)

				respBody, _ = ioutil.ReadAll(resp.Body)

				if bytes.Equal(expectedResponse, respBody) {
					break
				}

				time.Sleep(time.Millisecond * 100)
			}

			suite.Equal(string(expectedResponse), string(respBody))
		})
	}
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}

func id(i int) *int64 {
	x := int64(i)
	return &x
}

func doRequest(authHeader, path string, data interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8000"+os.Getenv("PATH_PREFIX")+path, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	return http.DefaultClient.Do(req)
}
