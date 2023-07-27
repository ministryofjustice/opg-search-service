package poadraftapplication

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestDraftApplication_Id(t *testing.T) {
	testId := int64(13)

	tests := []struct {
		scenario   string
		firm       DraftApplication
		expectedId interface{}
	}{
		{
			scenario:   "Blank DraftApplication",
			firm:       DraftApplication{},
			expectedId: int64(0),
		},
		{
			scenario: "DraftApplication with ID",
			firm: DraftApplication{
				ID: &testId,
			},
			expectedId: int64(13),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.firm.Id(), test.scenario)
	}
}

func TestDraftApplication_Validate(t *testing.T) {
	testId := int64(1)
	var noErrs []response.Error

	tests := []struct {
		scenario     string
		firm         DraftApplication
		expectedErrs []response.Error
	}{
		{
			"valid DraftApplication",
			DraftApplication{
				ID: &testId,
			},
			noErrs,
		},
		{
			"missing DraftApplication id",
			DraftApplication{},
			[]response.Error{
				{
					Name:        "id",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid DraftApplication id",
			DraftApplication{
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

func TestDraftApplication_IndexConfig(t *testing.T) {
	name, _, err := IndexConfig()

	assert.Nil(t, err)
	assert.Regexp(t, `[a-z]+_[a-z0-9]+`, name)
}
