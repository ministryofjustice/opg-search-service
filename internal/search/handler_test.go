package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/middleware"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockPrepareQuery struct {
	mock.Mock
}

func (m *mockPrepareQuery) Fn(req *Request) map[string]interface{} {
	return m.Called(req).Get(0).(map[string]interface{})
}

type SearchHandlerTestSuite struct {
	suite.Suite
	logger       *logrus.Logger
	esClient     *elasticsearch.MockESClient
	handler      *Handler
	recorder     *httptest.ResponseRecorder
	respBody     *string
	prepareQuery *mockPrepareQuery
}

func (suite *SearchHandlerTestSuite) SetupTest() {
	suite.logger, _ = test.NewNullLogger()
	suite.esClient = new(elasticsearch.MockESClient)
	suite.prepareQuery = &mockPrepareQuery{}
	suite.handler = NewHandler(suite.logger, suite.esClient, []string{"whatever"}, suite.prepareQuery.Fn)
	suite.recorder = httptest.NewRecorder()
}

func (suite *SearchHandlerTestSuite) ServeRequest(method string, url string, body string) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		suite.T().Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")
	suite.handler.ServeHTTP(suite.recorder, req.WithContext(ctx))
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
	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `errors":[{"name":"request","description":"unable to unmarshal JSON request"}]`)
}

func (suite *SearchHandlerTestSuite) Test_EmptyRequestBody() {
	reqBody := ""
	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"request body is empty"}]`)
}

func (suite *SearchHandlerTestSuite) Test_InvalidSearchRequestBody() {
	reqBody := `{"term":1}`
	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"unable to unmarshal JSON request"}]`)
}

func (suite *SearchHandlerTestSuite) Test_EmptySearchTerm() {
	reqBody := `{"term":"  "}`
	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"search term is required and cannot be empty"}]`)
}

func (suite *SearchHandlerTestSuite) Test_ESReturnsUnexpectedError() {
	reqBody := `{"term":"test"}`

	suite.prepareQuery.
		On("Fn", mock.Anything).
		Return(map[string]interface{}{})

	esCall := suite.esClient.On("Search", mock.Anything, mock.Anything, mock.Anything)
	esCall.Return(&elasticsearch.SearchResult{}, errors.New("test ES error"))

	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusInternalServerError, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"unexpected error from elasticsearch"}]`)
}

func (suite *SearchHandlerTestSuite) Test_ContextCancelledError() {
	reqBody := `{"term":"test"}`

	suite.prepareQuery.
		On("Fn", mock.Anything).
		Return(map[string]interface{}{})

	esCall := suite.esClient.On("Search", mock.Anything, mock.Anything, mock.Anything)
	esCall.Return(&elasticsearch.SearchResult{}, context.Canceled)

	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(499, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"search request was cancelled"}]`)
}

func (suite *SearchHandlerTestSuite) Test_SearchWithAllParameters() {
	reqBody := `{"term":"testTerm","from":10,"size":20,"person_types":["type1"]}`
	searchBody := map[string]interface{}{"whatever": nil}

	suite.prepareQuery.
		On("Fn", mock.Anything).
		Return(searchBody)

	result := &elasticsearch.SearchResult{
		Hits: []json.RawMessage{
			[]byte(`{"id":10,"firstname":"Test1","surname":"Test1"}`),
			[]byte(`{"id":20,"firstname":"Test2","surname":"Test2"}`),
		},
		Aggregations: map[string]map[string]int{
			"personType": {
				"firm": 2,
			},
		},
		Total:      2,
		TotalExact: true,
	}

	suite.esClient.
		On("Search", mock.Anything, []string{"whatever"}, searchBody).
		Return(result, nil)

	suite.ServeRequest(http.MethodPost, "", reqBody)

	expectedResponse := Response{
		Results:      result.Hits,
		Aggregations: result.Aggregations,
		Total: ResponseTotal{
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
