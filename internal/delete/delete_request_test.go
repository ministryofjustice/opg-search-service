package delete

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSearchRequestFromRequest(t *testing.T) {
	tests := []struct {
		scenario        string
		reqJson         string
		err             error
		expectedRequest *Request
	}{
		{
			"create from an empty request body",
			"",
			errors.New("request body is empty"),
			nil,
		},
		{
			"create from an unexpected request body",
			"not_a_json",
			errors.New("unable to unmarshal JSON request"),
			nil,
		},
		{
			"create from an invalid json",
			`{"uid":700013821983}`,
			errors.New("unable to unmarshal JSON request"),
			nil,
		},
		{
			"uid is required",
			`{"uid":""}`,
			errors.New("uid is required and cannot be empty"),
			nil,
		},
	}
	for _, test := range tests {
		req := http.Request{
			Body: ioutil.NopCloser(bytes.NewReader([]byte(test.reqJson))),
		}
		sr, err := parseDeleteRequest(&req)

		assert.Equal(t, test.err, err, test.scenario)
		assert.Equal(t, test.expectedRequest, sr, test.scenario)
	}
}
