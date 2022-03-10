package indices

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
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
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
	suite.handler = NewIndexHandler(suite.logger, suite.esClient, []string{"firm-test", "firm-new"})
	suite.router = mux.NewRouter().Methods(http.MethodPost).Subrouter()
	suite.router.Handle("/firms", suite.handler)
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
	suite.ServeRequest(http.MethodPost, "/firms", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), "Unable to unmarshal JSON request")
}

func (suite *IndexHandlerTestSuite) Test_EmptyRequestBody() {
	reqBody := ""
	suite.ServeRequest(http.MethodPost, "/firms", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), "Request body is empty")
}

func (suite *IndexHandlerTestSuite) Test_InvalidIndexRequestBody() {
	reqBody := `{"firms":[{"firmName":"test"}]}`
	suite.ServeRequest(http.MethodPost, "/firms", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `{"message":"Some fields have failed validation","errors":[{"name":"id","description":"field is empty"}]}`)
}

func (suite *IndexHandlerTestSuite) Test_Index() {
	reqBody := `{"firms":[{"id":13},{"id":14}]}`

	var id1, id2 int64 = 13, 14

	firstOp := elasticsearch.NewBulkOp("firm-test")
	firstOp.Index(13, Firm{ID: &id1})
	firstOp.Index(14, Firm{ID: &id2})
	secondOp := elasticsearch.NewBulkOp("firm-new")
	secondOp.Index(13, Firm{ID: &id1})
	secondOp.Index(14, Firm{ID: &id2})

	suite.esClient.
		On("DoBulk", firstOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 1}, errors.New("hmm")).
		Once()
	suite.esClient.
		On("DoBulk", secondOp).
		Return(elasticsearch.BulkResult{Successful: 2, Failed: 1}, errors.New("hey")).
		Once()

	suite.ServeRequest(http.MethodPost, "/firms", reqBody)

	suite.Equal(http.StatusAccepted, suite.RespCode())
	suite.Equal(`{"successful":4,"failed":2,"errors":["hmm","hey"]}`, suite.RespBody())
}

func TestIndexHandler(t *testing.T) {
	suite.Run(t, new(IndexHandlerTestSuite))
}

