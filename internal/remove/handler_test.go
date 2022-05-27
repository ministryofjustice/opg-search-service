package remove

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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DeleteHandlerTestSuite struct {
	suite.Suite
	logger   *logrus.Logger
	esClient *elasticsearch.MockESClient
	handler  *Handler
	recorder *httptest.ResponseRecorder
	respBody *string
}

func (suite *DeleteHandlerTestSuite) SetupTest() {
	suite.logger, _ = test.NewNullLogger()
	suite.esClient = new(elasticsearch.MockESClient)
	suite.handler = NewHandler(suite.logger, suite.esClient, []string{"whatever"})
	suite.recorder = httptest.NewRecorder()
}

const (
	varsKey = iota
)

func (suite *DeleteHandlerTestSuite) ServeRequest(method string, url string, vars map[string]string) {
	req, err := http.NewRequest(method, url, strings.NewReader(""))
	if err != nil {
		suite.T().Fatal(err)
	}

	ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")
	req = req.WithContext(ctx)
	req = mux.SetURLVars(req, vars)

	suite.handler.ServeHTTP(suite.recorder, req)
	suite.respBody = nil
}

func (suite *DeleteHandlerTestSuite) RespBody() string {
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

func (suite *DeleteHandlerTestSuite) RespCode() int {
	return suite.recorder.Result().StatusCode
}

func (suite *DeleteHandlerTestSuite) Test_MissingUid() {
	suite.ServeRequest(http.MethodDelete, "/persons/", map[string]string{})

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Equal(`{"message":"uid is required and cannot be empty","errors":[]}`+"\n", suite.RespBody())
}

func (suite *DeleteHandlerTestSuite) Test_ESReturnsUnexpectedError() {
	esCall := suite.esClient.On("Delete", mock.Anything, mock.Anything)
	esCall.Return(&elasticsearch.DeleteResult{}, errors.New("test ES error"))

	suite.ServeRequest(http.MethodDelete, "/persons/7000-9000-9201", map[string]string{"uid": "7000-9000-9201"})

	suite.Equal(http.StatusInternalServerError, suite.RespCode())
	suite.Equal(`{"message":"unexpected error from elasticsearch","errors":[]}`+"\n", suite.RespBody())
}

func (suite *DeleteHandlerTestSuite) Test_ESReturnsNoResults() {
	suite.esClient.
		On("Delete", []string{"whatever"}, mock.Anything).
		Return(&elasticsearch.DeleteResult{Total: 0}, nil)

	suite.ServeRequest(http.MethodDelete, "/persons/7000-9000-9201", map[string]string{"uid": "7000-9000-9201"})

	suite.Equal(http.StatusNotFound, suite.RespCode())
	suite.Equal(`{"message":"could not find document to delete","errors":[]}`+"\n", suite.RespBody())
}

func (suite *DeleteHandlerTestSuite) Test_ESReturnsMultipleResults() {
	suite.esClient.
		On("Delete", []string{"whatever"}, mock.Anything).
		Return(&elasticsearch.DeleteResult{Total: 2}, nil)

	suite.ServeRequest(http.MethodDelete, "/persons/7000-9000-9201", map[string]string{"uid": "7000-9000-9201"})

	suite.Equal(http.StatusInternalServerError, suite.RespCode())
	suite.Equal(`{"message":"deleted 2 documents matching query","errors":[]}`+"\n", suite.RespBody())
}

func (suite *DeleteHandlerTestSuite) Test_Delete() {
	deleteBody := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"uId": "7000-9000-9201",
			},
		},
	}

	result := &elasticsearch.DeleteResult{
		Total: 1,
	}

	suite.esClient.
		On("Delete", []string{"whatever"}, deleteBody).
		Return(result, nil)

	suite.ServeRequest(http.MethodDelete, "/persons/7000-9000-9201", map[string]string{"uid": "7000-9000-9201"})

	suite.Equal(http.StatusOK, suite.RespCode())
	suite.Equal("{}", suite.RespBody())
}

func TestDeleteHandler(t *testing.T) {
	suite.Run(t, new(DeleteHandlerTestSuite))
}
