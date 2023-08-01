package poadraftapplication

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestIndexRequest_Validate(t *testing.T) {
	testId := "M-789Q-P4DF-4UX3"
	var noErrs []response.Error

	tests := []struct {
		scenario string
		request IndexRequest
		expectErrs []response.Error
	}{
		{
			"valid request",
			IndexRequest{
				DraftApplications: []DraftApplication{
					{
						UID: &testId,
						Donor: DraftApplicationDonor{
							Name: "Vancep BVigliaon",
							Dob: "12/12/2000",
							Postcode: "X11 11X",
						},
					},
				},
			},
			noErrs,
		},
		{
			"missing DraftApplications",
			IndexRequest{},
			[]response.Error{
				{
					Name: "draftApplications",
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
					Name:        "draftApplications",
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
					Name:        "uId",
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
