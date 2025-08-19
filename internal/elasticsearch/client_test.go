package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/config"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	logrus_test "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var indexConfig = []byte("{json}")

type MockHttpClient struct {
	mock.Mock
}

func (m *MockHttpClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestClient_DoBulkIndex(t *testing.T) {
	tests := []struct {
		scenario           string
		esResponseError    error
		expectedStatusCode int
		expectedResponse   string
		expectedResult     BulkResult
		expectedError      string
		expectedLogs       []string
	}{
		{
			scenario:           "Document updated successfully",
			esResponseError:    nil,
			expectedStatusCode: 200,
			expectedResponse:   `{"errors":false,"items":[{"index":{"_id":"12","status":200}}]}`,
			expectedResult:     BulkResult{Successful: 1},
			expectedLogs:       []string{},
		},
		{
			scenario:           "Index request failure",
			esResponseError:    errors.New("some ES error"),
			expectedStatusCode: 500,
			expectedResponse:   "",
			expectedError:      "unable to process index request: some ES error",
			expectedLogs: []string{
				"some ES error",
			},
		},
		{
			scenario:           "Document failure",
			esResponseError:    nil,
			expectedStatusCode: 200,
			expectedResponse:   `{"errors":true,"items":[{"index":{"_id":"12","status":400}}]}`,
			expectedResult:     BulkResult{Failed: 1},
			expectedLogs:       []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			assert := assert.New(t)

			mc := new(MockHttpClient)

			l, hook := logrus_test.NewNullLogger()

			_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
			_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
			cfg, _ := config.LoadDefaultConfig(context.Background())
			c, err := NewClient(mc, l, &cfg)

			assert.IsType(&Client{}, c)
			assert.Nil(err)

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
					Body:       io.NopCloser(strings.NewReader(test.expectedResponse)),
				},
				test.esResponseError,
			)

			op := NewBulkOp("this")
			err = op.Index("1", map[string]string{"a": "b"})
			assert.Nil(err)

			result, err := c.DoBulk(context.Background(), op)

			if err != nil {
				assert.Equal(test.expectedError, err.Error())
			}
			assert.Equal(test.expectedResult, result)

			for i, logM := range test.expectedLogs {
				assert.Contains(hook.Entries[i].Message, logM)
			}
		})
	}
}

func TestClient_DoBulkIndexWithRetry(t *testing.T) {
	assert := assert.New(t)

	mc := new(MockHttpClient)

	l, _ := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	c, err := NewClient(mc, l, &cfg)

	assert.IsType(&Client{}, c)
	assert.Nil(err)

	mc.On("Do", mock.AnythingOfType("*http.Request")).
		Run(func(args mock.Arguments) {
			req := args[0].(*http.Request)
			assert.Equal(http.MethodPost, req.Method)
			assert.Equal(os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/this/_bulk", req.URL.String())
			reqBuf := new(bytes.Buffer)
			_, _ = reqBuf.ReadFrom(req.Body)
			assert.Equal(`{"index":{"_id":"1"}}
{"a":"b"}
`, reqBuf.String())
		}).
		Return(
			&http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			nil,
		).
		Once()

	mc.On("Do", mock.AnythingOfType("*http.Request")).
		Run(func(args mock.Arguments) {
			req := args[0].(*http.Request)
			assert.Equal(http.MethodPost, req.Method)
			assert.Equal(os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/this/_bulk", req.URL.String())
			reqBuf := new(bytes.Buffer)
			_, _ = reqBuf.ReadFrom(req.Body)
			assert.Equal(`{"index":{"_id":"1"}}
{"a":"b"}
`, reqBuf.String())
		}).
		Return(
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"errors":false,"items":[{"index":{"_id":"12","status":200}}]}`)),
			},
			nil,
		).
		Once()

	op := NewBulkOp("this")
	err = op.Index("1", map[string]string{"a": "b"})
	assert.Nil(err)

	result, err := c.DoBulk(context.Background(), op)

	assert.Nil(err)
	assert.Equal(BulkResult{Successful: 1}, result)
}

func TestClientCreateIndex(t *testing.T) {
	assert := assert.New(t)

	httpClient := &MockHttpClient{}
	l, hook := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodHead &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			data, _ := io.ReadAll(req.Body)

			return req.Method == http.MethodPut &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index" &&
				bytes.Equal(data, indexConfig)
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("test message"))}, nil).
		Once()

	err = client.CreateIndex(context.Background(), "test-index", indexConfig, false)
	assert.Nil(err)
	assert.Contains(hook.LastEntry().Message, "index 'test-index' created")
}

func TestClientCreateIndexWhenIndexExists(t *testing.T) {
	assert := assert.New(t)

	httpClient := &MockHttpClient{}
	l, hook := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodHead &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	err = client.CreateIndex(context.Background(), "test-index", indexConfig, false)
	assert.Nil(err)
	assert.Contains(hook.LastEntry().Message, "index 'test-index' already exists")
}

func TestClientCreateIndexWhenIndexExistsAndForced(t *testing.T) {
	assert := assert.New(t)

	httpClient := &MockHttpClient{}
	l, hook := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodHead &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodDelete &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			data, _ := io.ReadAll(req.Body)

			return req.Method == http.MethodPut &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index" &&
				bytes.Equal(data, indexConfig)
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("test message"))}, nil).
		Once()

	err = client.CreateIndex(context.Background(), "test-index", indexConfig, true)
	assert.Nil(err)
	assert.Contains(hook.LastEntry().Message, "index 'test-index' created")
}

func TestClientCreateIndexErrorIndexExists(t *testing.T) {
	assert := assert.New(t)

	httpClient := &MockHttpClient{}
	l, hook := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodHead &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	err = client.CreateIndex(context.Background(), "test-index", indexConfig, false)
	assert.NotNil(err)
	assert.Contains(hook.LastEntry().Message, "Checking index 'test-index' exists")
}

func TestClientCreateIndexErrorDeleteIndex(t *testing.T) {
	assert := assert.New(t)

	httpClient := &MockHttpClient{}
	l, hook := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodHead &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodDelete &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, errors.New("hey")).
		Once()

	err = client.CreateIndex(context.Background(), "test-index", indexConfig, true)
	assert.NotNil(err)
	assert.Contains(hook.LastEntry().Message, "Deleting index 'test-index'")
}

func TestClientCreateIndexErrorCreateIndex(t *testing.T) {
	assert := assert.New(t)

	httpClient := &MockHttpClient{}
	l, hook := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			return req.Method == http.MethodHead &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index"
		})).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil).
		Once()

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			data, _ := io.ReadAll(req.Body)

			return req.Method == http.MethodPut &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index" &&
				bytes.Equal(data, indexConfig)
		})).
		Return(&http.Response{StatusCode: http.StatusOK}, errors.New("hey")).
		Once()

	err = client.CreateIndex(context.Background(), "test-index", indexConfig, false)
	assert.NotNil(err)
	assert.Contains(hook.LastEntry().Message, "Creating index 'test-index'")
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
			esResponseMessage: `{"hits":{"hits":[{"_index":"person_foo1111","_source":{"id":1,"name":"test1"}},{"_index":"person_foo1111","_source":{"id":2,"name":"test1"}}]},"aggregations":{"personType":{"buckets":[{"key":"donor","doc_count":2}]}}}`,
			expectedError:     nil,
			expectedResult: &SearchResult{
				Hits: []json.RawMessage{
					[]byte(`{"_index":"person","id":1,"name":"test1"}`),
					[]byte(`{"_index":"person","id":2,"name":"test1"}`),
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

			_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
			_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
			cfg, _ := config.LoadDefaultConfig(context.Background())
			c, err := NewClient(mc, l, &cfg)

			assert.IsType(&Client{}, c)
			assert.Nil(err)

			reqBody := map[string]interface{}{
				"term": "test",
			}

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
					Body:       io.NopCloser(strings.NewReader(test.esResponseMessage)),
				},
				test.esResponseError,
			)

			result, err := c.Search(context.Background(), []string{"test-index"}, reqBody)

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

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	c, _ := NewClient(mc, l, &cfg)

	res, err := c.Search(context.Background(), []string{"test-index"}, map[string]interface{}{})

	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "missing protocol scheme")

	_ = os.Setenv("AWS_ELASTICSEARCH_ENDPOINT", oldESEndpoint)
}

func TestClient_Search_InvalidESRequestBody(t *testing.T) {
	mc := new(MockHttpClient)

	l, _ := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	c, _ := NewClient(mc, l, &cfg)

	esReqBody := map[string]interface{}{
		"term": func() {},
	}
	res, err := c.Search(context.Background(), []string{"test-index"}, esReqBody)

	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "json: unsupported type: func()")
}

func TestBulkOp(t *testing.T) {
	op := NewBulkOp("test")
	err := op.Index("1", map[string]interface{}{"a": 1})
	assert.Nil(t, err)
	err = op.Index("2", map[string]interface{}{"b": "c"})
	assert.Nil(t, err)
	err = op.Index("3", map[string]interface{}{"d": false})
	assert.Nil(t, err)

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

func TestDelete(t *testing.T) {
	tests := []struct {
		scenario          string
		esResponseError   error
		esResponseCode    int
		esResponseMessage string
		expectedError     error
		expectedResult    *DeleteResult
	}{
		{
			scenario:          "successful delete",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `{"total":1}`,
			expectedError:     nil,
			expectedResult: &DeleteResult{
				Total: 1,
			},
		},
		{
			scenario:          "delete HTTP fails",
			esResponseError:   nil,
			esResponseCode:    500,
			esResponseMessage: "failure",
			expectedError:     errors.New("delete request failed with status code 500 and response: \"failure\""),
			expectedResult:    nil,
		},
		{
			scenario:          "delete JSON fails",
			esResponseError:   nil,
			esResponseCode:    200,
			esResponseMessage: `bad body`,
			expectedError:     errors.New("error parsing the response body: invalid character 'b' looking for beginning of value"),
			expectedResult:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			assert := assert.New(t)

			mc := new(MockHttpClient)
			l, _ := logrus_test.NewNullLogger()

			_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
			_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
			cfg, _ := config.LoadDefaultConfig(context.Background())
			client, _ := NewClient(mc, l, &cfg)

			reqBody := map[string]interface{}{
				"query": map[string]interface{}{
					"match": map[string]interface{}{
						"uId": "7000-2837-9194",
					},
				},
			}

			mcCall := mc.On("Do", mock.AnythingOfType("*http.Request"))
			mcCall.RunFn = func(args mock.Arguments) {
				req := args[0].(*http.Request)
				assert.Equal(http.MethodPost, req.Method)
				assert.Equal(os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/test-index/_delete_by_query?conflicts=proceed", req.URL.String())
				reqBuf := new(bytes.Buffer)
				_, _ = reqBuf.ReadFrom(req.Body)
				assert.Equal(`{"query":{"match":{"uId":"7000-2837-9194"}}}`, strings.TrimSpace(reqBuf.String()))
			}

			mcCall.Return(
				&http.Response{
					StatusCode: test.esResponseCode,
					Body:       io.NopCloser(strings.NewReader(test.esResponseMessage)),
				},
				test.esResponseError,
			)

			result, err := client.Delete(context.Background(), []string{"test-index"}, reqBody)

			assert.Equal(test.expectedResult, result)

			if test.expectedError == nil {
				assert.Nil(err)
			} else {
				assert.Equal(test.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestClientSignRequest(t *testing.T) {
	assert := assert.New(t)

	AUTH_HEADER_PATTERN := `^AWS4-HMAC-SHA256 Credential=[^ ]+, SignedHeaders=content-type;host;x-amz-content-sha256;x-amz-date, Signature=[a-f0-9]+$`

	httpClient := &MockHttpClient{}
	l, _ := logrus_test.NewNullLogger()

	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client, err := NewClient(httpClient, l, &cfg)
	assert.Nil(err)

	//req.Header
	//AWS4-HMAC-SHA256 Credential=test/20250819/eu-west-1/es/aws4_request,
	//SignedHeaders=content-type;
	//host;x-amz-content-sha256;x-amz-date,
	//Signature=b4dfeb7772a38b198ae9c336c873ad0ebc81184f635ac409cb20083e7c9b2692

	//expected Header
	//^AWS4-HMAC-SHA256 Credential=[^ ]+,
	//SignedHeaders=content-type;
	//host;x-amz-date,
	//Signature=[a-f0-9]+$

	httpClient.
		On("Do", mock.MatchedBy(func(req *http.Request) bool {
			matched, _ := regexp.MatchString(AUTH_HEADER_PATTERN, req.Header["Authorization"][0])

			return req.Method == http.MethodGet &&
				req.URL.String() == os.Getenv("AWS_ELASTICSEARCH_ENDPOINT")+"/_healthcheck" &&
				matched
		})).
		Return(&http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil).Once()

	_, err = client.doRequest(context.Background(), http.MethodGet, "_healthcheck", bytes.NewReader([]byte{}), "application/json")
	assert.Nil(err)
}
