package search

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
			`{"term":10,"person_types":"all"}`,
			errors.New("unable to unmarshal JSON request"),
			nil,
		},
		{
			"white space is trimmed and search term validated",
			`{"term":"  "}`,
			errors.New("search term is required and cannot be empty"),
			nil,
		},
		{
			"create from a prepared request",
			`{"prepared":{"query":"some prepared query"}}`,
			nil,
			&Request{
				Prepared: map[string]interface{}{"query": "some prepared query"},
			},
		},
		{
			"created request is sanitised",
			`{"term":"R'ené_8 D’!Eath-Smi/the()","size":1,"from":2,"person_types":["firm","person"]}`,
			nil,
			&Request{
				Term: "R'ené_8 D’Eath-Smi/the",
				Size: 1,
				From: 2,
				PersonTypes: []string{
					"firm",
					"person",
				},
			},
		},
		{
			"created request can include indices to use for search",
			`{"term":"Vega","indices":["person","firm"]}`,
			nil,
			&Request{
				Term: "Vega",
				Indices: []string{"person", "firm"},
			},
		},
	}
	for _, test := range tests {
		req := http.Request{
			Body: ioutil.NopCloser(bytes.NewReader([]byte(test.reqJson))),
		}
		sr, err := parseSearchRequest(&req)

		assert.Equal(t, test.err, err, test.scenario)
		assert.Equal(t, test.expectedRequest, sr, test.scenario)
	}
}
