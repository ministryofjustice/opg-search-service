package deputy

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
)

func TestDeputy_Id(t *testing.T) {
	testId := int64(13)

	tests := []struct {
		scenario   string
		person     Deputy
		expectedId interface{}
	}{
		{
			scenario:   "Blank Deputy",
			person:     Deputy{},
			expectedId: int64(0),
		},
		{
			scenario: "Deputy with ID",
			person: Deputy{
				ID: &testId,
			},
			expectedId: int64(13),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.person.Id(), test.scenario)
	}
}

func TestDeputy_Validate(t *testing.T) {
	testId := int64(1)
	var noErrs []response.Error

	tests := []struct {
		scenario     string
		person       Deputy
		expectedErrs []response.Error
	}{
		{
			"valid deputy",
			Deputy{
				ID: &testId,
			},
			noErrs,
		},
		{
			"missing person id",
			Deputy{},
			[]response.Error{
				{
					Name:        "id",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid deputy id",
			Deputy{
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
		errs := test.person.Validate()
		assert.Equal(t, test.expectedErrs, errs, test.scenario)
	}
}

func TestDeputy_IndexConfig(t *testing.T) {
	name, _, err := IndexConfig()

	assert.Nil(t, err)
	assert.Regexp(t, `[a-z]+_[a-z0-9]+`, name)
}
