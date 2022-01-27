package cli

import (
	"errors"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"testing"

	"github.com/stretchr/testify/assert"
)

var indexConfig = []byte("{json}")

func TestCreateIndicesRun(t *testing.T) {
	const esErrorMessage = "some ES error"

	tests := []struct {
		scenario string
		force    bool
		error    error
	}{
		{
			scenario: "Index created successfully",
		},
		{
			scenario: "Error when creating index",
			error:    errors.New(esErrorMessage),
		},
		{
			scenario: "Force creating existing index",
			force:    true,
		},
		{
			scenario: "Error when force creating index",
			force:    true,
			error:    errors.New(esErrorMessage),
		},
	}
	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			esClient := new(elasticsearch.MockESClient)
			esClient.On("CreateIndex", person.AliasName, indexConfig, tc.force).Times(1).Return(nil).
				On("CreateIndex", "person-test", indexConfig, tc.force).Times(1).Return(tc.error)

			command := NewCreateIndices(esClient, "person-test", indexConfig)

			args := []string{}
			if tc.force {
				args = []string{"-force"}
			}

			err := command.Run(args)
			assert.Equal(t, tc.error, err)
		})
	}
}

func TestCreateIndicesRunErrorInFirst(t *testing.T) {
	const esErrorMessage = "some ES error"

	tests := []struct {
		scenario string
		force    bool
		error    error
	}{
		{
			scenario: "Error when creating index",
			error:    errors.New(esErrorMessage),
		},
		{
			scenario: "Error when force creating index",
			force:    true,
			error:    errors.New(esErrorMessage),
		},
	}
	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			esClient := new(elasticsearch.MockESClient)
			esClient.On("CreateIndex", "person-test", indexConfig, tc.force).Times(1).Return(tc.error)

			command := NewCreateIndices(esClient, "person-test", indexConfig)

			args := []string{}
			if tc.force {
				args = []string{"-force"}
			}

			err := command.Run(args)
			assert.Equal(t, tc.error, err)
		})
	}
}
