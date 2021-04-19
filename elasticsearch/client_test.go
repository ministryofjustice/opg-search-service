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
			scenario:           "Document created successfully",
			esResponseError:    nil,
			expectedStatusCode: 201,
			expectedMessage:    "Document created",
			expectedLogs: []string{
				"Indexing *elasticsearch.MockIndexable ID 6",
			},
		},
		{
			scenario:           "Document updated successfully",
			esResponseError:    nil,
			expectedStatusCode: 200,
			expectedMessage:    "Document updated",
			expectedLogs: []string{
				"Indexing *elasticsearch.MockIndexable ID 6",
			},
		},
		{
			scenario:           "Index request failure",
			esResponseError:    errors.New("some ES error"),
			expectedStatusCode: 500,
			expectedMessage:    "Unable to process document index request",
			expectedLogs: []string{
				"Indexing *elasticsearch.MockIndexable ID 6",
				"some ES error",
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

		for _, logM := range test.expectedLogs {
			assert.Contains(t, lBuf.String(), logM, test.scenario)
		}
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
	assert.Equal(t, "Unable to create document index request", ir.Message)

	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", oldESEndpoint)
}

func TestClient_CreateIndex(t *testing.T) {
	tests := []struct {
		scenario          string
		esResponseError   error
		esResponseCode    int
		esResponseMessage string
		expectedError     error
	}{
		{
			scenario:          "Index created successfully",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: "test message",
			expectedError:     nil,
		},
		{
			scenario:          "Create index request unexpected failure",
			esResponseError:   errors.New("some ES error"),
			esResponseCode:    500,
			esResponseMessage: "test message",
			expectedError:     errors.New("some ES error"),
		},
		{
			scenario:          "Create index request validation failure",
			esResponseError:   nil,
			esResponseCode:    400,
			esResponseMessage: "test message",
			expectedError:     errors.New(`index creation failed with status code 400 and response: "test message"`),
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
		mi.On("IndexName").Return("test-index").Times(2)
		mi.On("IndexConfig").Return(map[string]interface{}{"test": "test"}).Times(1)

		mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
		mcCall.RunFn = func(args mock.Arguments) {
			req := args[0].(*http.Request)
			assert.Equal(t, http.MethodPut, req.Method)
			assert.Equal(t, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index", req.URL.String())
			reqBuf := new(bytes.Buffer)
			_, _ = reqBuf.ReadFrom(req.Body)
			assert.Equal(t, `{"test":"test"}`, strings.TrimSpace(reqBuf.String()))
		}
		mcCall.Return(
			&http.Response{
				StatusCode: test.esResponseCode,
				Body:       ioutil.NopCloser(strings.NewReader(test.esResponseMessage)),
			},
			test.esResponseError,
		)

		result, err := c.CreateIndex(mi)

		assert.Contains(t, lBuf.String(), "Creating index 'test-index' for *elasticsearch.MockIndexable", test.scenario)
		assert.Equal(t, test.expectedError == nil, result, test.scenario)
		assert.Equal(t, test.expectedError, err, test.scenario)
	}
}

func TestClient_CreateIndex_MalformedEndpoint(t *testing.T) {
	oldESEndpoint := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")
	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", ":-:/-=")

	mc := new(MockHttpClient)

	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	c, _ := NewClient(mc, l)

	mi := new(MockIndexable)
	mi.On("IndexName").Return("test-index").Times(2)
	mi.On("IndexConfig").Return(map[string]interface{}{"test": "test"}).Times(1)

	res, err := c.CreateIndex(mi)

	assert.False(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")

	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", oldESEndpoint)
}

func TestClient_Search_InvalidIndexConfig(t *testing.T) {
	mc := new(MockHttpClient)

	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	c, _ := NewClient(mc, l)

	indexConfig := map[string]interface{}{
		"test": func() {},
	}

	mi := new(MockIndexable)
	mi.On("IndexName").Return("test-index").Times(2)
	mi.On("IndexConfig").Return(indexConfig).Times(1)

	res, err := c.CreateIndex(mi)

	assert.False(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "json: unsupported type: func()")
}

func TestClient_Search(t *testing.T) {
	tests := []struct {
		scenario          string
		esResponseError   error
		esResponseCode    int
		esResponseMessage string
		expectedError     error
		expectedResults   *[]string
	}{
		{
			scenario:          "Search returns matches",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `{"hits":{"hits":[{"_source":{"id":1,"name":"test1"}},{"_source":{"id":2,"name":"test1"}}]}}`,
			expectedError:     nil,
			expectedResults: &[]string{
				`{"id":1,"name":"test1"}`,
				`{"id":2,"name":"test1"}`,
			},
		},
		{
			scenario:          "Search does not return matches",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `{"hits":{"hits":[]}}`,
			expectedError:     nil,
			expectedResults:   &[]string{},
		},
		{
			scenario:          "Search request unexpected failure",
			esResponseError:   errors.New("some ES error"),
			esResponseCode:    500,
			esResponseMessage: "test message",
			expectedError:     errors.New("some ES error"),
			expectedResults:   nil,
		},
		{
			scenario:          "Search returns unexpected response body",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `<xml>not a json</xml>`,
			expectedError:     errors.New("error parsing the response body: invalid character '<' looking for beginning of value"),
			expectedResults:   nil,
		},
		{
			scenario:          "Search request validation failure",
			esResponseError:   nil,
			esResponseCode:    400,
			esResponseMessage: "test message",
			expectedError:     errors.New(`search request failed with status code 400 and response: "test message"`),
			expectedResults:   nil,
		},
	}

	for _, test := range tests {
		mc := new(MockHttpClient)

		lBuf := new(bytes.Buffer)
		l := log.New(lBuf, "", log.LstdFlags)

		c, err := NewClient(mc, l)

		assert.IsType(t, &Client{}, c, test.scenario)
		assert.Nil(t, err, test.scenario)

		reqBody := map[string]interface{}{
			"term": "test",
		}

		mi := new(MockIndexable)
		mi.On("IndexName").Return("test-index").Times(1)

		mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
		mcCall.RunFn = func(args mock.Arguments) {
			req := args[0].(*http.Request)
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index/_search", req.URL.String())
			reqBuf := new(bytes.Buffer)
			_, _ = reqBuf.ReadFrom(req.Body)
			assert.Equal(t, `{"term":"test"}`, strings.TrimSpace(reqBuf.String()))
		}
		mcCall.Return(
			&http.Response{
				StatusCode: test.esResponseCode,
				Body:       ioutil.NopCloser(strings.NewReader(test.esResponseMessage)),
			},
			test.esResponseError,
		)

		result, err := c.Search(reqBody, mi)

		assert.Equal(t, test.expectedResults, result, test.scenario)
		assert.Equal(t, test.expectedError, err, test.scenario)
	}
}

func TestClient_Search_MalformedEndpoint(t *testing.T) {
	oldESEndpoint := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")
	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", ":-:/-=")

	mc := new(MockHttpClient)

	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	c, _ := NewClient(mc, l)

	mi := new(MockIndexable)
	mi.On("IndexName").Return("test-index").Times(1)

	res, err := c.Search(map[string]interface{}{}, mi)

	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")

	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", oldESEndpoint)
}

func TestClient_Search_InvalidESRequestBody(t *testing.T) {
	mc := new(MockHttpClient)

	lBuf := new(bytes.Buffer)
	l := log.New(lBuf, "", log.LstdFlags)

	c, _ := NewClient(mc, l)

	mi := new(MockIndexable)
	mi.On("IndexName").Return("test-index").Times(1)

	esReqBody := map[string]interface{}{
		"term": func() {},
	}
	res, err := c.Search(esReqBody, mi)

	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "json: unsupported type: func()")
}
