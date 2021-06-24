package cli

import (
	"errors"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"os"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestNewCreateIndices(t *testing.T) {
	l, _ := test.NewNullLogger()
	ci := NewCreateIndices(l)
	assert.IsType(t, new(createIndices), ci)
	assert.IsType(t, new(elasticsearch.Client), ci.esClient)
	assert.Nil(t, ci.shouldRun)
	assert.IsType(t, os.Exit, ci.exit)
	assert.Same(t, l, ci.logger)
}

func TestCreateIndices_DefineFlags(t *testing.T) {
	l, _ := test.NewNullLogger()
	ci := NewCreateIndices(l)
	assert.Nil(t, ci.shouldRun)
	ci.DefineFlags()
	assert.False(t, *ci.shouldRun)
}

func TestCreateIndices_ShouldRun(t *testing.T) {
	tests := []struct {
		scenario  string
		shouldRun bool
	}{
		{
			scenario:  "Command should run",
			shouldRun: true,
		},
		{
			scenario:  "Command should not run",
			shouldRun: false,
		},
	}
	for _, test := range tests {
		ci := &createIndices{
			shouldRun: &test.shouldRun,
		}
		assert.Equal(t, test.shouldRun, ci.ShouldRun(), test.scenario)
	}
}

func TestCreateIndices_Run(t *testing.T) {
	const ESErrorMessage = "some ES error"

	tests := []struct {
		scenario     string
		force        bool
		esError      error
		esExists     bool
		esExistsErr  error
		esDeleteErr  error
		wantInLog    []string
		wantExitCode int
	}{
		{
			scenario:     "Index created successfully",
			esError:      nil,
			esExists:     false,
			esExistsErr:  nil,
			wantInLog:    []string{"Person index created successfully"},
			wantExitCode: 0,
		},
		{
			scenario:     "Error when creating index",
			esError:      errors.New(ESErrorMessage),
			esExists:     false,
			esExistsErr:  nil,
			wantInLog:    []string{ESErrorMessage},
			wantExitCode: 1,
		},
		{
			scenario:     "Index already exists",
			esError:      nil,
			esExists:     true,
			esExistsErr:  nil,
			wantInLog:    []string{"Person index already exists"},
			wantExitCode: 0,
		},
		{
			scenario:     "Error when checking if index exists",
			esError:      nil,
			esExists:     false,
			esExistsErr:  errors.New(ESErrorMessage),
			wantInLog:    []string{ESErrorMessage},
			wantExitCode: 1,
		},
		{
			scenario:     "Force delete existing index",
			force:        true,
			esError:      nil,
			esExists:     true,
			esExistsErr:  nil,
			wantInLog:    []string{"Person index already exists"},
			wantExitCode: 0,
		},
		{
			scenario:     "Error deleting existing index",
			force:        true,
			esError:      nil,
			esExists:     true,
			esExistsErr:  nil,
			esDeleteErr:  errors.New(ESErrorMessage),
			wantInLog:    []string{"Person index already exists", "Changes are forced, deleting old index", ESErrorMessage},
			wantExitCode: 1,
		},
	}
	for _, tc := range tests {
		l, hook := test.NewNullLogger()

		esClient := new(elasticsearch.MockESClient)
		esClient.On("IndexExists", person.Person{}).Times(1).Return(tc.esExists, tc.esExistsErr)
		esClient.On("DeleteIndex", person.Person{}).Times(1).Return(tc.esDeleteErr)
		esClient.On("CreateIndex", person.Person{}).Times(1).Return(tc.esError == nil, tc.esError)

		exitCode := 666
		ci := createIndices{
			logger:   l,
			esClient: esClient,
			force:    &tc.force,
			exit: func(code int) {
				exitCode = code
			},
		}

		ci.Run()

		for i, message := range tc.wantInLog {
			assert.Contains(t, hook.Entries[i].Message, message, tc.scenario)
		}
		assert.Equal(t, tc.wantExitCode, exitCode, tc.scenario)
	}
}
