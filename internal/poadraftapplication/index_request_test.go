package poadraftapplication

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
				DraftApplications: []DraftApplication{
					{ID: &testId},
				},
			},
			noErrs,
		},
		{
			"missing DraftApplications",
			IndexRequest{},
			[]response.Error{
				{
					Name: "entity",
					Description: "field is empty",
				},
			},
		},
		{
			"empty DraftApplications",
			IndexRequest{
				DraftApplications: []DraftApplication{},
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
				DraftApplications: []DraftApplication{
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
