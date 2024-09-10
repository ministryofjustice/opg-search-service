package response

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONErrors(t *testing.T) {
	tests := []struct {
		scenario string
		message  string
		errors   []Error
		code     int
		expected string
	}{
		{
			"Multiple errors",
			"test",
			[]Error{
				{
					Name:        "error1",
					Description: "descr1",
				},
				{
					Name:        "error2",
					Description: "descr2",
				},
			},
			500,
			`{"message":"test","errors":[{"name":"error1","description":"descr1"},{"name":"error2","description":"descr2"}]}` + "\n",
		},
		{
			"Blank errors",
			"",
			[]Error{},
			200,
			`{"message":"","errors":[]}` + "\n",
		},
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()
		WriteJSONErrors(rr, test.message, test.errors, test.code)

		r := rr.Result()
		b, _ := io.ReadAll(r.Body)

		assert.Equal(t, test.expected, string(b), test.scenario)
		assert.Equal(t, test.code, r.StatusCode, test.scenario)
	}
}

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		scenario string
		name     string
		descr    string
		code     int
		expected string
	}{
		{
			"Error fields populated",
			"err name",
			"err description",
			500,
			`{"message":"err name","errors":[{"name":"err name","description":"err description"}]}` + "\n",
		},
		{
			"Blank error",
			"",
			"",
			200,
			`{"message":"","errors":[{"name":"","description":""}]}` + "\n",
		},
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()
		WriteJSONError(rr, test.name, test.descr, test.code)

		r := rr.Result()
		b, _ := io.ReadAll(r.Body)

		assert.Equal(t, test.expected, string(b), test.scenario)
		assert.Equal(t, test.code, r.StatusCode, test.scenario)
	}
}
