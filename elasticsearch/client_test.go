package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	logrus_test "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestClient_DoBulkIndex(t *testing.T) {
	tests := []struct {
		scenario           string
		esResponseError    error
		expectedStatusCode int
		expectedResponse   string
		expectedResult     *BulkResult
		expectedLogs       []string
	}{
		{
			scenario:           "Document updated successfully",
			esResponseError:    nil,
			expectedStatusCode: 200,
			expectedResponse:   `{"errors":false}`,
			expectedResult:     &BulkResult{StatusCode: 200},
			expectedLogs:       []string{},
		},
		{
			scenario:           "Index request failure",
			esResponseError:    errors.New("some ES error"),
			expectedStatusCode: 500,
			expectedResponse:   "",
			expectedResult:     &BulkResult{StatusCode: 500, Message: "Unable to process document index request"},
			expectedLogs: []string{
				"some ES error",
			},
		},
		{
			scenario:           "Document failure",
			esResponseError:    nil,
			expectedStatusCode: 200,
			expectedResponse:   `{"errors":true,"items":[{"index":{"_id":"12","status":400}}]}`,
			expectedResult:     &BulkResult{StatusCode: 200, Results: []BulkResultItem{{ID: "12", StatusCode: 400}}},
			expectedLogs:       []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			assert := assert.New(t)

			mc := new(MockHttpClient)

			l, hook := logrus_test.NewNullLogger()

			c, err := NewClient(mc, l)

			assert.IsType(&Client{}, c)
			assert.Nil(err)

			mi := new(MockIndexable)
			mi.On("Id").Return(int64(6)).Times(3)
			mi.On("IndexName").Return("test-index").Times(1)
			mi.On("Json").Return("{\"test\":\"test\"}")

			mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
			mcCall.RunFn = func(args mock.Arguments) {
				req := args[0].(*http.Request)
				assert.Equal(http.MethodPost, req.Method)
				assert.Equal(os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/this/_bulk", req.URL.String())
				reqBuf := new(bytes.Buffer)
				_, _ = reqBuf.ReadFrom(req.Body)
				assert.Equal(`{"index":{"_id":"1"}}
{"a":"b"}
`, reqBuf.String())
			}
			mcCall.Return(
				&http.Response{
					StatusCode: test.expectedStatusCode,
					Body:       ioutil.NopCloser(strings.NewReader(test.expectedResponse)),
				},
				test.esResponseError,
			)

			op := NewBulkOp("this")
			op.Index(1, map[string]string{"a": "b"})

			result := c.DoBulk(op)

			assert.Equal(test.expectedResult, result)

			for i, logM := range test.expectedLogs {
				assert.Contains(hook.Entries[i].Message, logM)
			}
		})
	}
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
		t.Run(test.scenario, func(t *testing.T) {
			mc := new(MockHttpClient)

			l, hook := logrus_test.NewNullLogger()

			c, err := NewClient(mc, l)

			assert.IsType(t, &Client{}, c)
			assert.Nil(t, err)

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

			assert.Contains(t, hook.LastEntry().Message, "Creating index 'test-index' for *elasticsearch.MockIndexable")
			assert.Equal(t, test.expectedError == nil, result)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestClient_CreateIndex_MalformedEndpoint(t *testing.T) {
	oldESEndpoint := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")
	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", ":-:/-=")

	mc := new(MockHttpClient)

	l, _ := logrus_test.NewNullLogger()

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

func TestClient_DeleteIndex(t *testing.T) {
	tests := []struct {
		scenario        string
		esResponseError error
		esResponseCode  int
		expectedError   error
	}{
		{
			scenario:        "Index deleted successfully",
			esResponseError: nil,
			esResponseCode:  200,
			expectedError:   nil,
		},
		{
			scenario:        "Delete index request unexpected failure",
			esResponseError: errors.New("some ES error"),
			esResponseCode:  500,
			expectedError:   errors.New("some ES error"),
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			mc := new(MockHttpClient)

			l, hook := logrus_test.NewNullLogger()

			c, err := NewClient(mc, l)

			assert.IsType(t, &Client{}, c)
			assert.Nil(t, err)

			mi := new(MockIndexable)
			mi.On("IndexName").Return("test-index").Times(2)
			mi.On("IndexConfig").Return(map[string]interface{}{"test": "test"}).Times(1)

			mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
			mcCall.RunFn = func(args mock.Arguments) {
				req := args[0].(*http.Request)
				assert.Equal(t, http.MethodDelete, req.Method)
				assert.Equal(t, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index", req.URL.String())
			}
			mcCall.Return(
				&http.Response{
					StatusCode: test.esResponseCode,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				},
				test.esResponseError,
			)

			err = c.DeleteIndex(mi)

			assert.Contains(t, hook.LastEntry().Message, "Deleting index 'test-index' for *elasticsearch.MockIndexable")
			assert.Equal(t, test.expectedError, err)
		})
	}
}

func TestClient_DeleteIndex_MalformedEndpoint(t *testing.T) {
	oldESEndpoint := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")
	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", ":-:/-=")

	mc := new(MockHttpClient)

	l, _ := logrus_test.NewNullLogger()

	c, _ := NewClient(mc, l)

	mi := new(MockIndexable)
	mi.On("IndexName").Return("test-index").Times(2)
	mi.On("IndexConfig").Return(map[string]interface{}{"test": "test"}).Times(1)

	err := c.DeleteIndex(mi)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")

	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", oldESEndpoint)
}

func TestClient_Search_InvalidIndexConfig(t *testing.T) {
	mc := new(MockHttpClient)

	l, _ := logrus_test.NewNullLogger()

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
		expectedResult    *SearchResult
	}{
		{
			scenario:          "Search returns matches",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `{"hits":{"hits":[{"_source":{"id":1,"name":"test1"}},{"_source":{"id":2,"name":"test1"}}]},"aggregations":{"personType":{"buckets":[{"key":"donor","doc_count":2}]}}}`,
			expectedError:     nil,
			expectedResult: &SearchResult{
				Hits: []json.RawMessage{
					[]byte(`{"id":1,"name":"test1"}`),
					[]byte(`{"id":2,"name":"test1"}`),
				},
				Aggregations: map[string]map[string]int{
					"personType": {
						"donor": 2,
					},
				},
			},
		},
		{
			scenario:          "Search does not return matches",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `{"hits":{"hits":[]}}`,
			expectedError:     nil,
			expectedResult: &SearchResult{
				Hits:         []json.RawMessage{},
				Aggregations: map[string]map[string]int{},
			},
		},
		{
			scenario:          "Search request unexpected failure",
			esResponseError:   errors.New("some ES error"),
			esResponseCode:    500,
			esResponseMessage: "test message",
			expectedError:     errors.New("some ES error"),
			expectedResult:    nil,
		},
		{
			scenario:          "Search returns unexpected response body",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `<xml>not a json</xml>`,
			expectedError:     errors.New("error parsing the response body: invalid character '<' looking for beginning of value"),
			expectedResult:    nil,
		},
		{
			scenario:          "Search request validation failure",
			esResponseError:   nil,
			esResponseCode:    400,
			esResponseMessage: "test message",
			expectedError:     errors.New(`search request failed with status code 400 and response: "test message"`),
			expectedResult:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			assert := assert.New(t)
			mc := new(MockHttpClient)

			l, _ := logrus_test.NewNullLogger()

			c, err := NewClient(mc, l)

			assert.IsType(&Client{}, c)
			assert.Nil(err)

			reqBody := map[string]interface{}{
				"term": "test",
			}

			mi := new(MockIndexable)
			mi.On("IndexName").Return("test-index").Times(1)

			mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
			mcCall.RunFn = func(args mock.Arguments) {
				req := args[0].(*http.Request)
				assert.Equal(http.MethodPost, req.Method)
				assert.Equal(os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index/_search", req.URL.String())
				reqBuf := new(bytes.Buffer)
				_, _ = reqBuf.ReadFrom(req.Body)
				assert.Equal(`{"term":"test"}`, strings.TrimSpace(reqBuf.String()))
			}
			mcCall.Return(
				&http.Response{
					StatusCode: test.esResponseCode,
					Body:       ioutil.NopCloser(strings.NewReader(test.esResponseMessage)),
				},
				test.esResponseError,
			)

			result, err := c.Search(reqBody, mi)

			assert.Equal(test.expectedResult, result)
			if test.expectedError == nil {
				assert.Nil(err)
			} else {
				assert.Equal(test.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestClient_Search_MalformedEndpoint(t *testing.T) {
	oldESEndpoint := os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")
	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", ":-:/-=")

	mc := new(MockHttpClient)

	l, _ := logrus_test.NewNullLogger()

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

	l, _ := logrus_test.NewNullLogger()

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

func TestClient_IndexExists(t *testing.T) {
	tests := []struct {
		scenario        string
		esResponseError error
		esResponseCode  int
		wantExists      bool
		wantError       error
	}{
		{
			scenario:        "Index exists",
			esResponseError: nil,
			esResponseCode:  200,
			wantExists:      true,
			wantError:       nil,
		},
		{
			scenario:        "Index does not exist",
			esResponseError: nil,
			esResponseCode:  404,
			wantExists:      false,
			wantError:       nil,
		},
		{
			scenario:        "Unexpected failure",
			esResponseError: errors.New("some ES error"),
			esResponseCode:  500,
			wantExists:      false,
			wantError:       errors.New("some ES error"),
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			mc := new(MockHttpClient)

			l, hook := logrus_test.NewNullLogger()

			c, err := NewClient(mc, l)

			assert.IsType(t, &Client{}, c)
			assert.Nil(t, err)

			mi := new(MockIndexable)
			mi.On("IndexName").Return("test-index").Times(2)

			mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
			mcCall.RunFn = func(args mock.Arguments) {
				req := args[0].(*http.Request)
				assert.Equal(t, http.MethodHead, req.Method)
				assert.Equal(t, os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index", req.URL.String())
			}
			mcCall.Return(
				&http.Response{
					StatusCode: test.esResponseCode,
				},
				test.esResponseError,
			)

			exists, err := c.IndexExists(mi)

			assert.Contains(t, hook.LastEntry().Message, "Checking index 'test-index' exists")
			assert.Equal(t, test.wantExists, exists)
			assert.Equal(t, test.wantError, err)
		})
	}
}

func TestBulkOp(t *testing.T) {
	op := NewBulkOp("test")
	op.Index(1, map[string]interface{}{"a": 1})
	op.Index(2, map[string]interface{}{"b": "c"})
	op.Index(3, map[string]interface{}{"d": false})

	expected := `{"index":{"_id":"1"}}
{"a":1}
{"index":{"_id":"2"}}
{"b":"c"}
{"index":{"_id":"3"}}
{"d":false}
`

	assert.True(t, strings.HasSuffix(expected, "\n"), "ensure the final newline is not removed")

	result := op.buf.String()
	assert.Equal(t, expected, result)
}
