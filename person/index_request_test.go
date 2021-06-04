package person

import (
	"github.com/stretchr/testify/assert"
	"opg-search-service/response"
	"testing"
)

func TestIndexRequest_Validate(t *testing.T) {
	testId := int64(1)
	var noErrs []response.Error

	tests := []struct {
		scenario   string
		request    IndexRequest
		expectErrs []response.Error
	}{
		{
			"valid request",
			IndexRequest{
				Persons: []Person{
					{ID: &testId},
				},
			},
			noErrs,
		},
		{
			"missing persons",
			IndexRequest{},
			[]response.Error{
				{
					Name:        "persons",
					Description: "field is empty",
				},
			},
		},
		{
			"empty persons",
			IndexRequest{
				Persons: []Person{},
			},
			[]response.Error{
				{
					Name:        "persons",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid person error propagates",
			IndexRequest{
				Persons: []Person{
					{},
				},
			},
			[]response.Error{
				{
					Name:        "id",
					Description: "field is empty",
				},
			},
		},
	}
	for _, test := range tests {
		errs := test.request.Validate()
		assert.Equal(t, errs, test.expectErrs, test.scenario)
	}
}
