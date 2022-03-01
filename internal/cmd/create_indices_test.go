package cmd

import (
	"errors"
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
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
			esClient.
				On("CreateIndex", "person-test", indexConfig, tc.force).Times(1).Return(tc.error)

			if tc.error == nil {
				esClient.On("ResolveAlias", "person").Times(1).Return("", nil)
			}

			command := NewCreateIndices(esClient, "person-test", indexConfig)

			args := []string{}
			if tc.force {
				args = []string{"-force"}
			}

			err := command.Run(args)
			assert.Equal(t, tc.error, err)
			esClient.AssertExpectations(t)
		})
	}
}

func TestCreateIndicesRunCreateAlias(t *testing.T) {
	esClient := new(elasticsearch.MockESClient)
	esClient.
		On("CreateIndex", "person-test", indexConfig, true).Times(1).Return(nil).
		On("ResolveAlias", "person").Times(1).Return("", elasticsearch.ErrAliasMissing).
		On("CreateAlias", "person", "person-test").Times(1).Return(nil)

	command := NewCreateIndices(esClient, "person-test", indexConfig)

	args := []string{"-force"}

	err := command.Run(args)
	assert.Equal(t, nil, err)
	esClient.AssertExpectations(t)
}

func TestCreateIndicesRunCreateAliasFails(t *testing.T) {
	creationErr := errors.New("error creating alias")

	esClient := new(elasticsearch.MockESClient)
	esClient.
		On("CreateIndex", "person-test", indexConfig, true).Times(1).Return(nil).
		On("ResolveAlias", "person").Times(1).Return("", elasticsearch.ErrAliasMissing).
		On("CreateAlias", "person", "person-test").Times(1).Return(creationErr)

	command := NewCreateIndices(esClient, "person-test", indexConfig)

	args := []string{"-force"}

	err := command.Run(args)
	assert.Equal(t, creationErr, err)
	esClient.AssertExpectations(t)
}

func TestCreateIndicesRunResolveAliasFails(t *testing.T) {
	resolveErr := errors.New("error creating alias")

	esClient := new(elasticsearch.MockESClient)
	esClient.
		On("CreateIndex", "person-test", indexConfig, true).Times(1).Return(nil).
		On("ResolveAlias", "person").Times(1).Return("", resolveErr)

	command := NewCreateIndices(esClient, "person-test", indexConfig)

	args := []string{"-force"}

	err := command.Run(args)
	assert.Equal(t, resolveErr, err)
	esClient.AssertExpectations(t)
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
			esClient.AssertExpectations(t)
		})
	}
}
