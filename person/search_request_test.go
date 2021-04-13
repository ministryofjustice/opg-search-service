package person

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestCreateSearchRequestFromRequest(t *testing.T) {
	tests := []struct {
		scenario        string
		reqJson         string
		err             error
		expectedRequest *searchRequest
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
			`{"term":10,"person_types":"all"}`,
			errors.New("unable to unmarshal JSON request"),
			nil,
		},
		{
			"created request is sanitised",
			`{"term":"René D’!Eath-Smi/the()","size":1,"from":2,"person_types":["tall","short"]}`,
			nil,
			&searchRequest{
				Term: "René D’Eath-Smithe",
				Size: 1,
				From: 2,
				PersonTypes: []string{
					"tall",
					"short",
				},
			},
		},
	}
	for _, test := range tests {
		req := http.Request{
			Body: ioutil.NopCloser(bytes.NewReader([]byte(test.reqJson))),
		}
		sr, err := CreateSearchRequestFromRequest(&req)

		assert.Equal(t, test.err, err, test.scenario)
		assert.Equal(t, test.expectedRequest, sr, test.scenario)
	}
}
