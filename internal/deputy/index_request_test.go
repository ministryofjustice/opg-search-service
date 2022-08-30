package deputy

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
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
				Deputies: []Deputy{
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
				Deputies: []Deputy{},
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
				Deputies: []Deputy{
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
