package person

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"opg-search-service/elasticsearch"
	"opg-search-service/middleware"
	"opg-search-service/response"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type SearchHandlerTestSuite struct {
	suite.Suite
	logger   *log.Logger
	esClient *elasticsearch.MockESClient
	handler  *SearchHandler
	router   *mux.Router
	recorder *httptest.ResponseRecorder
	respBody *string
}

func (suite *SearchHandlerTestSuite) SetupTest() {
	buf := new(bytes.Buffer)
	suite.logger = log.New(buf, "test", log.LstdFlags)
	suite.esClient = new(elasticsearch.MockESClient)
	suite.handler = &SearchHandler{
		logger: suite.logger,
		es:     suite.esClient,
	}
	suite.router = mux.NewRouter().Methods(http.MethodPost).Subrouter()
	suite.router.Handle("/persons/search", suite.handler)
	suite.recorder = httptest.NewRecorder()
}

func (suite *SearchHandlerTestSuite) ServeRequest(method string, url string, body string) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		suite.T().Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")
	suite.router.ServeHTTP(suite.recorder, req.WithContext(ctx))
	suite.respBody = nil
}

func (suite *SearchHandlerTestSuite) RespBody() string {
	if suite.respBody != nil {
		return *suite.respBody
	}
	res := suite.recorder.Result()
	bodyBuf := new(bytes.Buffer)
	_, _ = bodyBuf.ReadFrom(res.Body)
	respBody := bodyBuf.String()

	suite.respBody = &respBody
	return *suite.respBody
}

func (suite *SearchHandlerTestSuite) RespCode() int {
	return suite.recorder.Result().StatusCode
}

func (suite *SearchHandlerTestSuite) Test_InvalidJSONRequestBody() {
	reqBody := ".\\|{."
	suite.ServeRequest(http.MethodPost, "/persons/search", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `errors":[{"name":"request","description":"unable to unmarshal JSON request"}]`)
}

func (suite *SearchHandlerTestSuite) Test_EmptyRequestBody() {
	reqBody := ""
	suite.ServeRequest(http.MethodPost, "/persons/search", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"request body is empty"}]`)
}

func (suite *SearchHandlerTestSuite) Test_InvalidSearchRequestBody() {
	reqBody := `{"term":1}`
	suite.ServeRequest(http.MethodPost, "/persons/search", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"unable to unmarshal JSON request"}]`)
}

func (suite *SearchHandlerTestSuite) Test_EmptySearchTerm() {
	reqBody := `{"term":"  "}`
	suite.ServeRequest(http.MethodPost, "/persons/search", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"search term is required and cannot be empty"}]`)
}

func (suite *SearchHandlerTestSuite) Test_ESReturnsUnexpectedError() {
	reqBody := `{"term":"test"}`

	esCall := suite.esClient.On("Search", mock.Anything, mock.Anything)
	esCall.Return(&elasticsearch.SearchResult{}, errors.New("test ES error"))

	suite.ServeRequest(http.MethodPost, "/persons/search", reqBody)

	suite.Equal(http.StatusInternalServerError, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"Person search caused an unexpected error"}]`)
}

func (suite *SearchHandlerTestSuite) Test_SearchWithAllParameters() {
	reqBody := `{"term":"testTerm","from":10,"size":20,"person_types":["type1","type2"]}`

	esCall := suite.esClient.On("Search", mock.Anything, mock.AnythingOfType("Person"))
	esCall.RunFn = func(args mock.Arguments) {
		esReqBody := args[0].(map[string]interface{})
		dataType := args[1].(Person)
		expectedEsReqBody := map[string]interface{}{
			"size": 20,
			"from": 10,
			"sort": map[string]interface{}{
				"surname": map[string]string{
					"order": "asc",
				},
			},
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": map[string]interface{}{
						"simple_query_string": map[string]interface{}{
							"query": "testTerm",
							"fields": []string{
								"searchable",
								"caseRecNumber",
							},
							"default_operator": "AND",
						},
					},
					"filter": []interface{}{
						map[string]interface{}{
							"term": map[string]string{
								"personType": "type1",
							},
						},
						map[string]interface{}{
							"term": map[string]string{
								"personType": "type2",
							},
						},
					},
				},
			},
			"aggs": map[string]interface{}{
				"personType": map[string]interface{}{
					"terms": map[string]string{
						"field": "personType",
					},
				},
			},
		}
		suite.Equal(expectedEsReqBody, esReqBody)
		suite.IsType(Person{}, dataType)
	}

	result := &elasticsearch.SearchResult{
		Hits: []json.RawMessage{
			[]byte(`{"id":10,"firstname":"Test1","surname":"Test1"}`),
			[]byte(`{"id":20,"firstname":"Test2","surname":"Test2"}`),
		},
		Aggregations: map[string]map[string]int{
			"personType": {
				"attorney": 1,
				"donor":    1,
			},
		},
		Total:      2,
		TotalExact: true,
	}

	esCall.Return(result, nil)

	suite.ServeRequest(http.MethodPost, "/persons/search", reqBody)

	expectedResponse := response.SearchResponse{
		Results:      result.Hits,
		Aggregations: result.Aggregations,
		Total: response.Total{
			Count: 2,
			Exact: true,
		},
	}
	expectedJsonResponse, _ := json.Marshal(expectedResponse)

	suite.Equal(http.StatusOK, suite.RespCode())
	suite.Equal(string(expectedJsonResponse), suite.RespBody())
}

func TestSearchHandler(t *testing.T) {
	suite.Run(t, new(SearchHandlerTestSuite))
}

func TestNewSearchHandler(t *testing.T) {
	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	sh, err := NewSearchHandler(l)

	assert.Nil(t, err)
	assert.IsType(t, &SearchHandler{}, sh)
}
