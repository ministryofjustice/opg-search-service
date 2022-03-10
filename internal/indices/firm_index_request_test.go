package indices

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
				Firms: []Firm{
					{ID: &testId},
				},
			},
			noErrs,
		},
		{
			"missing firms",
			IndexRequest{},
			[]response.Error{
				{
					Name:        "entity",
					Description: "field is empty",
				},
			},
		},
		{
			"empty firms",
			IndexRequest{
				Firms: []Firm{},
			},
			[]response.Error{
				{
					Name:        "entity",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid firm error propagates",
			IndexRequest{
				Firms: []Firm{
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

