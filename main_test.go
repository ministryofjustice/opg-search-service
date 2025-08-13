package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/cmd"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/ministryofjustice/opg-search-service/internal/search"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
)

// add an _index field to marshalled JSON object objs
func toJSONWithIndex(objs []byte, index string) ([]byte, error) {
	var tmp map[string]interface{}

	err := json.Unmarshal(objs, &tmp)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal JSON for hit: %w", err)
	}

	tmp["_index"] = index

	result, err := json.Marshal(tmp)
	if err != nil {
		return nil, fmt.Errorf("unable to add _index to JSON: %w", err)
	}

	return result, nil
}

func personToJSON(person person.Person) ([]byte, error) {
	objs, _ := json.Marshal(person)
	return toJSONWithIndex(objs, "person")
}

func firmToJSON(firm firm.Firm) ([]byte, error) {
	objs, _ := json.Marshal(firm)
	return toJSONWithIndex(objs, "firm")
}

type EndToEndTestSuite struct {
	suite.Suite
	testPeople []person.Person
	testFirms  []firm.Firm
	esClient   *elasticsearch.Client
	authHeader string
}

func makeToken() string {
	exp := time.Now().AddDate(0, 1, 0).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session-data": "Test.McTestFace@mail.com",
		"iat":          time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
		"exp":          exp,
	})
	tokenString, err := token.SignedString([]byte("MyTestSecret"))
	if err != nil {
		log.Fatal("Could not make test token")
	}
	return tokenString
}

func (suite *EndToEndTestSuite) SetupSuite() {
	_ = os.Setenv("ENVIRONMENT", "local")
	_ = os.Setenv("JWT_SECRET", "MyTestSecret")
	_ = os.Setenv("USER_HASH_SALT", "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0")
	_ = os.Setenv("ENVIRONMENT", "local")
	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	_ = os.Setenv("AWS_SECRETS_MANAGER_ENDPOINT", "http://localstack:4566")

	logger, _ := test.NewNullLogger()
	httpClient := &http.Client{}
	ctx := context.Background()
	cfg, _ := awsConfig(ctx)
	suite.esClient, _ = elasticsearch.NewClient(httpClient, logger, cfg)

	suite.authHeader = "Bearer " + makeToken()

	suite.testPeople = []person.Person{
		{
			ID:         i64(0),
			Firstname:  "John0",
			Surname:    "Doe0",
			Persontype: "Type0",
			Dob:        "01/02/1990",
			Addresses: []person.PersonAddress{{
				Postcode: "NG1 2CD",
			}},
		},
		{
			ID:           i64(1),
			Firstname:    "John1",
			Surname:      "Doe1",
			Persontype:   "Type1",
			DeputyNumber: i64(12345),
			Dob:          "20/03/1987",
			Addresses: []person.PersonAddress{{
				Postcode: "NG1 1AB",
			}},
			Cases: []person.PersonCase{{
				OnlineLpaId: "ABCDEFGH",
			}},
		},
	}

	suite.testFirms = []firm.Firm{
		{
			ID:           i64(0),
			FirmName:     "Firm1",
			FirmNumber:   "1",
			Persontype:   "Firm",
			Email:        "test@test.com",
			AddressLine1: "Address Line 1",
			AddressLine2: "Address Line 2",
			AddressLine3: "Address Line 3",
			Town:         "Town",
			County:       "County",
			Postcode:     "PO2 CDE",
			Phonenumber:  "0123 456 789",
		},
		{
			ID:           i64(1),
			FirmName:     "Firm2",
			FirmNumber:   "2",
			Persontype:   "Firm",
			Email:        "test@test.com",
			AddressLine1: "Address Line 1",
			AddressLine2: "Address Line 2",
			AddressLine3: "Address Line 3",
			Town:         "Town",
			County:       "County",
			Postcode:     "PO2 CDE",
			Phonenumber:  "0123 456 789",
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

	personIndexConfig := cmd.NewIndexConfig(person.IndexConfig, person.AliasName, logger)
	firmIndexConfig := cmd.NewIndexConfig(firm.IndexConfig, firm.AliasName, logger)

	exists, err := suite.esClient.IndexExists(ctx, personIndexConfig.Name)
	suite.False(exists, "Person index should not exist at this point")
	suite.Nil(err)

	existsFirmIndex, err := suite.esClient.IndexExists(ctx, firmIndexConfig.Name)
	suite.False(existsFirmIndex, "Firm index should not exist at this point")
	suite.Nil(err)

	// wait up to 5 seconds for the app to start
	retries := 5
	for i := 1; i <= retries; i++ {
		conn, err := net.DialTimeout("tcp", "localhost:8000", time.Second)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		exists, err = suite.esClient.IndexExists(ctx, personIndexConfig.Name)
		suite.True(exists, "Person index should exist at this point")
		suite.Nil(err)

		existsFirmIndex, err = suite.esClient.IndexExists(ctx, firmIndexConfig.Name)
		suite.True(existsFirmIndex, "Firm index should exist at this point")
		suite.Nil(err)

		conn.Close() //nolint:errcheck,gosec // no need to check connection close error in tests
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
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	suite.Equal(http.StatusAccepted, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)

	suite.Equal(`{"successful":2,"failed":0}`, string(data))

	testCases := []struct {
		scenario         string
		term             string
		expectedResponse func() search.Response
	}{
		{
			scenario: "search by surname",
			term:     suite.testPeople[1].Surname,
			expectedResponse: func() search.Response {
				hit, _ := personToJSON(suite.testPeople[1])

				return search.Response{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type1": 1,
						},
					},
					Total: search.ResponseTotal{
						Count: 1,
						Exact: true,
					},
				}
			},
		},
		{
			scenario: "search by dob",
			term:     "01/02/1990",
			expectedResponse: func() search.Response {
				hit, _ := personToJSON(suite.testPeople[0])

				return search.Response{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type0": 1,
						},
					},
					Total: search.ResponseTotal{
						Count: 1,
						Exact: true,
					},
				}
			},
		},
		{
			scenario: "search by postcode",
			term:     "NG1 2CD",
			expectedResponse: func() search.Response {
				hit, _ := personToJSON(suite.testPeople[0])

				return search.Response{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type0": 1,
						},
					},
					Total: search.ResponseTotal{
						Count: 1,
						Exact: true,
					},
				}
			},
		},
		{
			scenario: "search by a-ref",
			term:     suite.testPeople[1].Cases[0].OnlineLpaId,
			expectedResponse: func() search.Response {
				hit, _ := personToJSON(suite.testPeople[1])

				return search.Response{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type1": 1,
						},
					},
					Total: search.ResponseTotal{
						Count: 1,
						Exact: true,
					},
				}
			},
		},
		{
			scenario: "search by deputy number",
			term:     strconv.FormatInt(*suite.testPeople[1].DeputyNumber, 10),
			expectedResponse: func() search.Response {
				hit, _ := personToJSON(suite.testPeople[1])

				return search.Response{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Type1": 1,
						},
					},
					Total: search.ResponseTotal{
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

				respBody, _ = io.ReadAll(resp.Body)

				if bytes.Equal(expectedResponse, respBody) {
					break
				}

				time.Sleep(time.Millisecond * 100)
			}

			suite.Equal(string(expectedResponse), string(respBody))
		})
	}
}

func (suite *EndToEndTestSuite) TestIndexAndSearchFirm() {
	resp, err := doRequest(suite.authHeader, "/firms", firm.IndexRequest{Firms: suite.testFirms})
	if err != nil {
		suite.Fail("Error indexing firm", err)
	}
	defer resp.Body.Close() //nolint:errcheck // no need to check error when closing body

	suite.Equal(http.StatusAccepted, resp.StatusCode)

	data, _ := io.ReadAll(resp.Body)

	suite.Equal(`{"successful":2,"failed":0}`, string(data))

	testCases := []struct {
		scenario         string
		term             string
		expectedResponse func() search.Response
	}{
		{
			scenario: "search by firmname",
			term:     suite.testFirms[1].FirmName,
			expectedResponse: func() search.Response {
				hit, _ := firmToJSON(suite.testFirms[1])

				return search.Response{
					Results: []json.RawMessage{hit},
					Aggregations: map[string]map[string]int{
						"personType": {
							"Firm": 1,
						},
					},
					Total: search.ResponseTotal{
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
				resp, err := doRequest(suite.authHeader, "/firms/search", map[string]string{"term": tc.term})
				if err != nil {
					suite.Fail("Error searching for a firm", err)
				}

				suite.Equal(http.StatusOK, resp.StatusCode)

				respBody, _ = io.ReadAll(resp.Body)

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
	if testing.Short() {
		t.Skip("skipping end to end tests")
		return
	}

	suite.Run(t, new(EndToEndTestSuite))
}

func i64(i int) *int64 {
	x := int64(i)
	return &x
}

func doRequest(authHeader, path string, data interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8000"+os.Getenv("PATH_PREFIX")+path, &buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	return http.DefaultClient.Do(req)
}
