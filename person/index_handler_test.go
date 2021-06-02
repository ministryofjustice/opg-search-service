package person

import (
	"bytes"
	"context"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"net/http/httptest"
	"opg-search-service/elasticsearch"
	"opg-search-service/middleware"
	"strings"
	"testing"
)

type IndexHandlerTestSuite struct {
	suite.Suite
	logger   *log.Logger
	esClient *elasticsearch.MockESClient
	handler  *IndexHandler
	router   *mux.Router
	recorder *httptest.ResponseRecorder
	respBody *string
}

func (suite *IndexHandlerTestSuite) SetupTest() {
	buf := new(bytes.Buffer)
	suite.logger = log.New(buf, "test", log.LstdFlags)
	suite.esClient = new(elasticsearch.MockESClient)
	suite.handler = &IndexHandler{
		logger: suite.logger,
		es:     suite.esClient,
	}
	suite.router = mux.NewRouter().Methods(http.MethodPost).Subrouter()
	suite.router.Handle("/persons", suite.handler)
	suite.recorder = httptest.NewRecorder()
}

func (suite *IndexHandlerTestSuite) ServeRequest(method string, url string, body string) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		suite.T().Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")
	suite.router.ServeHTTP(suite.recorder, req.WithContext(ctx))
	suite.respBody = nil
}

func (suite *IndexHandlerTestSuite) RespBody() string {
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

func (suite *IndexHandlerTestSuite) RespCode() int {
	return suite.recorder.Result().StatusCode
}

func (suite *IndexHandlerTestSuite) Test_InvalidJSONRequestBody() {
	reqBody := ".\\|{."
	suite.ServeRequest(http.MethodPost, "/persons", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), "Unable to unmarshal JSON request")
}

func (suite *IndexHandlerTestSuite) Test_EmptyRequestBody() {
	reqBody := ""
	suite.ServeRequest(http.MethodPost, "/persons", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), "Request body is empty")
}

func (suite *IndexHandlerTestSuite) Test_InvalidIndexRequestBody() {
	reqBody := `{"persons":[{"uid":"13"}]}`
	suite.ServeRequest(http.MethodPost, "/persons", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `{"message":"Some fields have failed validation","errors":[{"name":"id","description":"field is empty"}]}`)
}

func (suite *IndexHandlerTestSuite) Test_IndexSingle() {
	reqBody := `{"persons":[{"id":13}]}`

	id := int64(13)
	esCall := suite.esClient.On("Index", mock.AnythingOfType("Person"))
	esCall.RunFn = func(args mock.Arguments) {
		i := args[0].(Person)
		suite.Equal(Person{
			ID: &id,
		}, i)
	}
	esCall.Return(&elasticsearch.IndexResult{
		Id:         id,
		StatusCode: 200,
		Message:    "test success",
	})

	suite.ServeRequest(http.MethodPost, "/persons", reqBody)

	suite.Equal(http.StatusAccepted, suite.RespCode())
	suite.Equal(`{"results":[{"id":13,"statusCode":200,"message":"test success"}]}`, suite.RespBody())
}

func (suite *IndexHandlerTestSuite) Test_IndexMultiple() {
	reqBody := `{"persons":[{"id":13},{"id":14}]}`

	ids := [2]int64{13, 14}

	esCall := suite.esClient.On("Index", mock.AnythingOfType("Person"))
	esCall.RunFn = func(args mock.Arguments) {
		i := args[0].(Person)
		suite.Equal(Person{
			ID: &ids[0],
		}, i)
	}
	esCall.Return(&elasticsearch.IndexResult{
		Id:         ids[0],
		StatusCode: 200,
		Message:    "test success",
	}).Once()

	esCall = suite.esClient.On("Index", mock.AnythingOfType("Person"))
	esCall.RunFn = func(args mock.Arguments) {
		i := args[0].(Person)
		suite.Equal(Person{
			ID: &ids[1],
		}, i)
	}
	esCall.Return(&elasticsearch.IndexResult{
		Id:         ids[1],
		StatusCode: 200,
		Message:    "test success",
	}).Once()

	suite.ServeRequest(http.MethodPost, "/persons", reqBody)

	suite.Equal(http.StatusAccepted, suite.RespCode())
	suite.Equal(`{"results":[{"id":13,"statusCode":200,"message":"test success"},{"id":14,"statusCode":200,"message":"test success"}]}`, suite.RespBody())
}

func TestIndexHandler(t *testing.T) {
	suite.Run(t, new(IndexHandlerTestSuite))
}

func TestNewIndexHandler(t *testing.T) {
	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	ih, err := NewIndexHandler(l)

	assert.Nil(t, err)
	assert.IsType(t, &IndexHandler{}, ih)
}
