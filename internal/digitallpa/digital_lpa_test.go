package digitallpa

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestDigitalLpa_Id(t *testing.T) {
	testId := "M-789Q-P4DF-4UX3"

	tests := []struct {
		scenario   string
		digitalLpa DigitalLpa
		expectedId interface{}
	}{
		{
			scenario: "Digital LPA with UID",
			digitalLpa: DigitalLpa{
				Uid: testId,
			},
			expectedId: testId,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.digitalLpa.Id(), test.scenario)
	}
}

func TestDigitalLpa_Validate(t *testing.T) {
	testUid := "M-789Q-P4DF-4UX3"
	var noErrs []response.Error

	tests := []struct {
		scenario     string
		digitalLpa   DigitalLpa
		expectedErrs []response.Error
	}{
		{
			"valid digital Lpa",
			DigitalLpa{
				Uid: testUid,
			},
			noErrs,
		},
		{
			"empty digital Lpa uid",
			DigitalLpa{
				Uid: "",
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
		errs := test.digitalLpa.Validate()
		assert.Equal(t, test.expectedErrs, errs, test.scenario)
	}
}

func TestDigitalLpa_IndexConfig(t *testing.T) {
	name, _, err := IndexConfig()
	assert.Nil(t, err)
	assert.Regexp(t, `[a-z]+_[a-z0-9]+`, name)
}
