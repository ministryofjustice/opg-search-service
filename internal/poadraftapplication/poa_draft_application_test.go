package poadraftapplication

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestDraftApplication_Id(t *testing.T) {
	testId := "M-789Q-P4DF-4UX3"

	tests := []struct {
		scenario string
		draftApplication DraftApplication
		expectedId interface{}
	}{
		{
			scenario: "Blank DraftApplication",
			draftApplication: DraftApplication{},
			expectedId: "0",
		},
		{
			scenario: "DraftApplication with UID",
			draftApplication: DraftApplication{
				UID: &testId,
			},
			expectedId: testId,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.draftApplication.Id(), test.scenario)
	}
}

func TestDraftApplication_Validate(t *testing.T) {
	testUid := "M-789Q-P4DF-4UX3"
	var noErrs []response.Error

	tests := []struct {
		scenario string
		draftApplication DraftApplication
		expectedErrs []response.Error
	}{
		{
			"valid DraftApplication",
			DraftApplication{
				UID: &testUid,
			},
			noErrs,
		},
		{
			"missing DraftApplication uid",
			DraftApplication{},
			[]response.Error{
				{
					Name:        "uId",
					Description: "field is empty",
				},
			},
		},
		{
			"nil DraftApplication uid",
			DraftApplication{
				UID: nil,
			},
			[]response.Error{
				{
					Name:        "uId",
					Description: "field is empty",
				},
			},
		},
		{
			"empty DraftApplication uid",
			DraftApplication{
				UID: new(string),
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
		errs := test.draftApplication.Validate()
		assert.Equal(t, test.expectedErrs, errs, test.scenario)
	}
}

func TestDraftApplication_IndexConfig(t *testing.T) {
	name, _, err := IndexConfig()
	assert.Nil(t, err)
	assert.Regexp(t, `[a-z]+_[a-z0-9]+`, name)
}
