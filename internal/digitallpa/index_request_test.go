package digitallpa

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestIndexRequest_Validate(t *testing.T) {
	testId := "M-789Q-P4DF-4UX3"
	var noErrs []response.Error

	tests := []struct {
		scenario   string
		request    IndexRequest
		expectErrs []response.Error
	}{
		{
			"valid request",
			IndexRequest{
				DigitalLpaApplications: []DigitalLpa{
					{
						Uid:     testId,
						LpaType: "PA",
						Donor: Donor{
							Person: Person{
								Firstnames: "Vancep",
								Surname:    "BVigliaon",
								Address: Address{
									Line1:    "18 Candle Street",
									Line2:    "",
									Line3:    "",
									Postcode: "X11 11X",
								},
							},
							Dob: "12/12/2000",
						},
					},
				},
			},
			noErrs,
		},
		{
			"missing Digital Lpa",
			IndexRequest{},
			[]response.Error{
				{
					Name:        "digitalLpaApplications",
					Description: "field is empty",
				},
			},
		},
		{
			"empty digital Lpa",
			IndexRequest{
				DigitalLpaApplications: []DigitalLpa{},
			},
			[]response.Error{
				{
					Name:        "digitalLpaApplications",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid digital Lpa error propagates",
			IndexRequest{
				DigitalLpaApplications: []DigitalLpa{
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
