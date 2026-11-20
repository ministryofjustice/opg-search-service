package poas_person

import (
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIndexRequest(t *testing.T) {
	request, err := ParseIndexRequest([]byte(`{"persons":[{"id":1,"deputyNumber":12,"deputyType":"professional","cases":[{"onlineLpaId":"A123","batchId":"B123"}]}]}`))
	require.NoError(t, err)
	require.Empty(t, request.Validate())
	items := request.Items()
	require.Len(t, items, 1)
	assert.Equal(t, Person{ID: i64(1), Cases: []PersonCase{{OnlineLpaId: "A123", Batchid: "B123"}}}, items[0])
}

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
				Persons: []Person{
					{ID: &testId},
				},
			},
			noErrs,
		},
		{
			"missing persons",
			IndexRequest{},
			[]response.Error{
				{
					Name:        "persons",
					Description: "field is empty",
				},
			},
		},
		{
			"empty persons",
			IndexRequest{
				Persons: []Person{},
			},
			[]response.Error{
				{
					Name:        "persons",
					Description: "field is empty",
				},
			},
		},
		{
			"invalid person error propagates",
			IndexRequest{
				Persons: []Person{
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
