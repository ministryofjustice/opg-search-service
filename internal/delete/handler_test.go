package delete

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

func (suite *DeleteHandlerTestSuite) ServeRequest(method string, url string, body string) {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		suite.T().Fatal(err)
	}
	ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")
	suite.handler.ServeHTTP(suite.recorder, req.WithContext(ctx))
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

func (suite *DeleteHandlerTestSuite) Test_InvalidJSONRequestBody() {
	reqBody := ".\\|{."
	suite.ServeRequest(http.MethodDelete, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `errors":[{"name":"request","description":"unable to unmarshal JSON request"}]`)
}

func (suite *DeleteHandlerTestSuite) Test_EmptyRequestBody() {
	reqBody := ""
	suite.ServeRequest(http.MethodDelete, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"request body is empty"}]`)
}

func (suite *DeleteHandlerTestSuite) Test_InvalidDeleteRequestBody() {
	reqBody := `{"uid":1}`
	suite.ServeRequest(http.MethodDelete, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"unable to unmarshal JSON request"}]`)
}

func (suite *DeleteHandlerTestSuite) Test_EmptyUid() {
	reqBody := `{"uid":""}`
	suite.ServeRequest(http.MethodDelete, "", reqBody)

	suite.Equal(http.StatusBadRequest, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"uid is required and cannot be empty"}]`)
}

func (suite *DeleteHandlerTestSuite) Test_ESReturnsUnexpectedError() {
	reqBody := `{"uid":"7000-9000-9201"}`

	esCall := suite.esClient.On("Delete", mock.Anything, mock.Anything)
	esCall.Return(&elasticsearch.DeleteResult{}, errors.New("test ES error"))

	suite.ServeRequest(http.MethodDelete, "", reqBody)

	suite.Equal(http.StatusInternalServerError, suite.RespCode())
	suite.Contains(suite.RespBody(), `"errors":[{"name":"request","description":"unexpected error from elasticsearch"}]`)
}

func (suite *DeleteHandlerTestSuite) Test_Delete() {
	reqBody := `{"uid":"7000-9000-9201"}`
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

	suite.ServeRequest(http.MethodDelete, "", reqBody)

	expectedResponse := Response{
		Total: 1,
	}
	expectedJsonResponse, _ := json.Marshal(expectedResponse)

	suite.Equal(http.StatusOK, suite.RespCode())
	suite.Equal(string(expectedJsonResponse), suite.RespBody())
}

func TestDeleteHandler(t *testing.T) {
	suite.Run(t, new(DeleteHandlerTestSuite))
}
