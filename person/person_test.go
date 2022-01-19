package person

import (
	"opg-search-service/response"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerson_Id(t *testing.T) {
	testId := int64(13)

	tests := []struct {
		scenario   string
		person     Person
		expectedId interface{}
	}{
		{
			scenario:   "Blank Person",
			person:     Person{},
			expectedId: int64(0),
		},
		{
			scenario: "Person with ID",
			person: Person{
				ID: &testId,
			},
			expectedId: int64(13),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.person.Id(), test.scenario)
	}
}

func TestPerson_Validate(t *testing.T) {
	testId := int64(1)
	var noErrs []response.Error

	tests := []struct {
		scenario     string
		person       Person
		expectedErrs []response.Error
	}{
		{
			"valid person",
			Person{
				ID: &testId,
			},
			noErrs,
		},
		{
			"missing person id",
			Person{},
			[]response.Error{
				{
					Name:        "id",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid person id",
			Person{
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
