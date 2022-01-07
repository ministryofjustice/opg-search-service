package cli

import (
	"errors"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestCreateIndicesRun(t *testing.T) {
	const esErrorMessage = "some ES error"

	tests := []struct {
		scenario    string
		force       bool
		esError     error
		esExists    bool
		esExistsErr error
		esDeleteErr error
		wantInLog   []string
		wantErr     error
	}{
		{
			scenario:    "Index created successfully",
			esError:     nil,
			esExists:    false,
			esExistsErr: nil,
			wantInLog:   []string{"Person index created successfully"},
			wantErr:     nil,
		},
		{
			scenario:    "Error when creating index",
			esError:     errors.New(esErrorMessage),
			esExists:    false,
			esExistsErr: nil,
			wantInLog:   []string{},
			wantErr:     errors.New(esErrorMessage),
		},
		{
			scenario:    "Index already exists",
			esError:     nil,
			esExists:    true,
			esExistsErr: nil,
			wantInLog:   []string{"Person index already exists"},
			wantErr:     nil,
		},
		{
			scenario:    "Error when checking if index exists",
			esError:     nil,
			esExists:    false,
			esExistsErr: errors.New(esErrorMessage),
			wantInLog:   []string{},
			wantErr:     errors.New(esErrorMessage),
		},
		{
			scenario:    "Force delete existing index",
			force:       true,
			esError:     nil,
			esExists:    true,
			esExistsErr: nil,
			wantInLog:   []string{"Person index already exists"},
			wantErr:     nil,
		},
		{
			scenario:    "Error deleting existing index",
			force:       true,
			esError:     nil,
			esExists:    true,
			esExistsErr: nil,
			esDeleteErr: errors.New(esErrorMessage),
			wantInLog:   []string{"Person index already exists", "Changes are forced, deleting old index"},
			wantErr:     errors.New(esErrorMessage),
		},
	}
	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			l, hook := test.NewNullLogger()

			esClient := new(elasticsearch.MockESClient)
			esClient.On("IndexExists", person.Person{}).Times(1).Return(tc.esExists, tc.esExistsErr)
			esClient.On("DeleteIndex", person.Person{}).Times(1).Return(tc.esDeleteErr)
			esClient.On("CreateIndex", person.Person{}).Times(1).Return(tc.esError == nil, tc.esError)

			ci := createIndices{
				logger:   l,
				esClient: esClient,
				force:    &tc.force,
			}

			err := ci.Run([]string{})
			assert.Equal(t, tc.wantErr, err)

			for i, message := range tc.wantInLog {
				assert.Contains(t, message, hook.Entries[i].Message)
			}
		})
	}
}
