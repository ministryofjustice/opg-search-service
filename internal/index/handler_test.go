package index

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/middleware"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
)

type HandlerTestSuite struct {
	suite.Suite
	logger            *logrus.Logger
	esClient          *elasticsearch.MockESClient
	handler           *Handler
	router            *mux.Router
	recorder          *httptest.ResponseRecorder
	respBody          *string
	parserValidatable Validatable
	parserError       error
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.logger, _ = test.NewNullLogger()
	suite.esClient = &elasticsearch.MockESClient{}
	suite.parserValidatable = nil
	suite.parserError = nil
	suite.handler = NewHandler(suite.logger, suite.esClient, []string{"whatever-test", "whatever-new"}, func(body []byte) (Validatable, error) {
		return suite.parserValidatable, suite.parserError
	})
	suite.recorder = httptest.NewRecorder()
}

func (suite *HandlerTestSuite) ServeRequest(method string, url string, body string) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		suite.T().Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")
	suite.handler.ServeHTTP(suite.recorder, req.WithContext(ctx))
	suite.respBody = nil
}

func (suite *HandlerTestSuite) RespBody() string {
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

func (suite *HandlerTestSuite) RespCode() int {
	return suite.recorder.Result().StatusCode
}

func (suite *HandlerTestSuite) Test_ParserError() {
	suite.parserError = errors.New("parse error")
	suite.ServeRequest(http.MethodPost, "", "whatever")

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), "Unable to unmarshal JSON request")
}

func (suite *HandlerTestSuite) Test_EmptyRequestBody() {
	suite.ServeRequest(http.MethodPost, "", "")

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), "Request body is empty")
}

func (suite *HandlerTestSuite) Test_InvalidIndexRequestBody() {
	suite.parserValidatable = mockValidatable{error: []response.Error{{
		Name:        "id",
		Description: "field is empty",
	}}}

	reqBody := `{"whatevers":[{"uid":"13"}]}`
	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `{"message":"Some fields have failed validation","errors":[{"name":"id","description":"field is empty"}]}`)
}

func (suite *HandlerTestSuite) Test_Index() {
	suite.parserValidatable = mockValidatable{
		items: []Indexable{
			mockIndexable{id: 13},
			mockIndexable{id: 14},
		},
	}

	reqBody := `{"whatevers":[{"id":13},{"id":14}]}`

	firstOp := elasticsearch.NewBulkOp("whatever-test")
	firstOp.Index(13, mockIndexable{id: 13})
	firstOp.Index(14, mockIndexable{id: 14})
	secondOp := elasticsearch.NewBulkOp("whatever-new")
	secondOp.Index(13, mockIndexable{id: 13})
	secondOp.Index(14, mockIndexable{id: 14})

	suite.esClient.
		On("DoBulk", firstOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 1}, errors.New("hmm")).
		Once()
	suite.esClient.
		On("DoBulk", secondOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 1}, errors.New("hey")).
		Once()

	suite.ServeRequest(http.MethodPost, "", reqBody)

	suite.Equal(http.StatusAccepted, suite.RespCode())
	suite.Equal(`{"successful":4,"failed":2,"errors":["hmm","hey"]}`, suite.RespBody())
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
