package elasticsearch

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

type MockIndexable struct {
	mock.Mock
}

func (m *MockIndexable) Id() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockIndexable) IndexName() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockIndexable) Json() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockIndexable) IndexConfig() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func TestClient_Index(t *testing.T) {
	tests := []struct {
		scenario           string
		esResponseError    error
		expectedStatusCode int
		expectedMessage    string
		expectedLogs       []string
	}{
		{
			scenario:           "Index created successfully",
			esResponseError:    nil,
			expectedStatusCode: 201,
			expectedMessage:    "Index created",
			expectedLogs: []string{
				"Indexing MockIndexable ID 6",
			},
		},
		{
			scenario:           "Index updated successfully",
			esResponseError:    nil,
			expectedStatusCode: 200,
			expectedMessage:    "Index updated",
			expectedLogs: []string{
				"Indexing MockIndexable ID 6",
			},
		},
		{
			scenario:           "Index request failure",
			esResponseError:    errors.New("some ES error"),
			expectedStatusCode: 500,
			expectedMessage:    "Unable to process index request",
			expectedLogs: []string{
				"Indexing MockIndexable ID 6",
				"Unable to process index request",
			},
		},
	}

	for _, test := range tests {
		mc := new(MockHttpClient)

		lBuf := new(bytes.Buffer)
		l := log.New(lBuf, "", log.LstdFlags)

		c, err := NewClient(mc, l)

		assert.IsType(t, &Client{}, c, test.scenario)
		assert.Nil(t, err, test.scenario)

		mi := new(MockIndexable)
		mi.On("Id").Return(int64(6)).Times(3)
		mi.On("IndexName").Return("test-index").Times(1)
		mi.On("Json").Return("{\"test\":\"test\"}")

		mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
		mcCall.RunFn = func(args mock.Arguments) {
			req := args[0].(*http.Request)
			assert.Equal(t, http.MethodPut, req.Method)
			assert.Equal(t, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index/_doc/6", req.URL.String())
			reqBuf := new(bytes.Buffer)
			_, _ = reqBuf.ReadFrom(req.Body)
			assert.Equal(t, "{\"test\":\"test\"}", reqBuf.String())
		}
		mcCall.Return(
			&http.Response{
				StatusCode: test.expectedStatusCode,
				Body:       ioutil.NopCloser(strings.NewReader(test.expectedMessage)),
			},
			test.esResponseError,
		)

		ir := c.Index(mi)

		assert.Equal(t, int64(6), ir.Id, test.scenario)
		assert.Equal(t, test.expectedStatusCode, ir.StatusCode, test.scenario)
		assert.Equal(t, test.expectedMessage, ir.Message, test.scenario)
	}
}

func TestClient_Index_MalformedEndpoint(t *testing.T) {
	oldESEndpoint := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")
	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", ":-:/-=")

	mc := new(MockHttpClient)

	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	c, err := NewClient(mc, l)

	assert.IsType(t, &Client{}, c)
	assert.Nil(t, err)

	mi := new(MockIndexable)
	mi.On("Id").Return(int64(6)).Times(3)
	mi.On("IndexName").Return("test-index").Times(1)
	mi.On("Json").Return("{\"test\":\"test\"}")

	ir := c.Index(mi)

	assert.Equal(t, int64(6), ir.Id)
	assert.Equal(t, http.StatusInternalServerError, ir.StatusCode)
	assert.Equal(t, "Unable to create index request", ir.Message)

	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", oldESEndpoint)
}
