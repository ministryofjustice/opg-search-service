package person

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"opg-search-service/response"
	"testing"
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
				Normalizeduid: &testId,
			},
			expectedId: int64(13),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.expectedId, test.person.Id(), test.scenario)
	}
}

func TestPerson_IndexName(t *testing.T) {
	p := Person{}
	assert.Equal(t, "person", p.IndexName())
}

func TestPerson_Json(t *testing.T) {
	p := Person{}
	res, _ := json.Marshal(p)
	assert.Equal(t, string(res), p.Json())
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
				Normalizeduid: &testId,
			},
			noErrs,
		},
		{
			"missing person id",
			Person{},
			[]response.Error{
				{
					Name:        "normalizedUid",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid person id",
			Person{
				Normalizeduid: nil,
			},
			[]response.Error{
				{
					Name:        "normalizedUid",
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
