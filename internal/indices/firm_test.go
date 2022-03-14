package indices

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestFirm_Id(t *testing.T) {
	testId := int64(13)

	tests := []struct {
		scenario   string
		firm     Firm
		expectedId interface{}
	}{
		{
			scenario:   "Blank Firm",
			firm:     Firm{},
			expectedId: int64(0),
		},
		{
			scenario: "Firm with ID",
			firm: Firm{
				ID: &testId,
			},
			expectedId: int64(13),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.firm.Id(), test.scenario)
	}
}

func TestFirm_Validate(t *testing.T) {
	testId := int64(1)
	var noErrs []response.Error

	tests := []struct {
		scenario     string
		firm       Firm
		expectedErrs []response.Error
	}{
		{
			"valid firm",
			Firm{
				ID: &testId,
			},
			noErrs,
		},
		{
			"missing firm id",
			Firm{},
			[]response.Error{
				{
					Name:        "id",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid person id",
			Firm{
				ID: nil,
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
		errs := test.firm.Validate()
		assert.Equal(t, test.expectedErrs, errs, test.scenario)
	}
}

