package person

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"opg-search-service/elasticsearch"
	"opg-search-service/middleware"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IndexHandlerTestSuite struct {
	suite.Suite
	logger   *logrus.Logger
	esClient *elasticsearch.MockESClient
	handler  *IndexHandler
	router   *mux.Router
	recorder *httptest.ResponseRecorder
	respBody *string
}

func (suite *IndexHandlerTestSuite) SetupTest() {
	suite.logger, _ = test.NewNullLogger()
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

func (suite *IndexHandlerTestSuite) Test_Index() {
	reqBody := `{"persons":[{"id":13},{"id":14}]}`

	esCall := suite.esClient.On("DoBulk", mock.AnythingOfType("*elasticsearch.BulkOp"))
	esCall.RunFn = func(args mock.Arguments) {
		actual := args[0].(*elasticsearch.BulkOp)

		var id1, id2 int64 = 13, 14

		expected := elasticsearch.NewBulkOp("person")
		expected.Index(13, Person{ID: &id1})
		expected.Index(14, Person{ID: &id2})

		suite.Equal(expected, actual)
	}
	esCall.Return([]elasticsearch.IndexResult{{
		StatusCode: 200,
		Id:         13,
		Message:    "test success",
	}, {
		StatusCode: 200,
		Id:         14,
		Message:    "test success",
	}}).Once()

	suite.ServeRequest(http.MethodPost, "/persons", reqBody)

	suite.Equal(http.StatusAccepted, suite.RespCode())
	suite.Equal(`{"results":[{"id":13,"statusCode":200,"message":"test success"},{"id":14,"statusCode":200,"message":"test success"}]}`, suite.RespBody())
}

func (suite *IndexHandlerTestSuite) Test_IndexBatchSize() {
	reqBody := `{"persons":[{"id":13},{"id":14}]}`

	var id1, id2 int64 = 13, 14

	suite.esClient.On("DoBulk", mock.AnythingOfType("*elasticsearch.BulkOp")).
		Run(func(args mock.Arguments) {
			actual := args[0].(*elasticsearch.BulkOp)

			expected := elasticsearch.NewBulkOp("person")

			expected.Index(13, Person{ID: &id1})
			suite.Equal(expected, actual)
		}).
		Return([]elasticsearch.IndexResult{{
			StatusCode: 200,
			Id:         13,
			Message:    "test success",
		}}).
		Once()

	suite.esClient.On("DoBulk", mock.AnythingOfType("*elasticsearch.BulkOp")).
		Run(func(args mock.Arguments) {
			actual := args[0].(*elasticsearch.BulkOp)

			expected := elasticsearch.NewBulkOp("person")

			expected.Index(14, Person{ID: &id2})
			suite.Equal(expected, actual)
		}).
		Return([]elasticsearch.IndexResult{{
			StatusCode: 200,
			Id:         14,
			Message:    "test success",
		}}).
		Once()

	suite.ServeRequest(http.MethodPost, "/persons?batchSize=1", reqBody)

	suite.Equal(http.StatusAccepted, suite.RespCode())
	suite.Equal(`{"results":[{"id":13,"statusCode":200,"message":"test success"},{"id":14,"statusCode":200,"message":"test success"}]}`, suite.RespBody())
}

func TestIndexHandler(t *testing.T) {
	suite.Run(t, new(IndexHandlerTestSuite))
}

func TestNewIndexHandler(t *testing.T) {
	l, _ := test.NewNullLogger()

	ih, err := NewIndexHandler(l)

	assert.Nil(t, err)
	assert.IsType(t, &IndexHandler{}, ih)
}
