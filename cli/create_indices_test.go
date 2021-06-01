package cli

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"os"
	"testing"
)

func TestNewCreateIndices(t *testing.T) {
	l := new(log.Logger)
	ci := NewCreateIndices(l)
	assert.IsType(t, new(createIndices), ci)
	assert.IsType(t, new(elasticsearch.Client), ci.esClient)
	assert.Nil(t, ci.shouldRun)
	assert.IsType(t, os.Exit, ci.exit)
	assert.Same(t, l, ci.logger)
}

func TestCreateIndices_DefineFlags(t *testing.T) {
	ci := NewCreateIndices(new(log.Logger))
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
	tests := []struct {
		scenario     string
		esError      error
		wantInLog    string
		wantExitCode int
	}{
		{
			scenario:     "Index created successfully",
			esError:      nil,
			wantInLog:    "Person index created successfully",
			wantExitCode: 0,
		},
		{
			scenario:     "Error when creating index",
			esError:      errors.New("some ES error"),
			wantInLog:    "some ES error",
			wantExitCode: 1,
		},
	}
	for _, test := range tests {
		lBuf := new(bytes.Buffer)
		l := log.New(lBuf, "", log.LstdFlags)

		esClient := new(elasticsearch.MockESClient)
		esClient.On("CreateIndex", person.Person{}).Times(1).Return(test.esError == nil, test.esError)

		exitCode := 666
		ci := createIndices{
			logger:   l,
			esClient: esClient,
			exit: func(code int) {
				exitCode = code
			},
		}

		ci.Run()

		assert.Contains(t, lBuf.String(), test.wantInLog, test.scenario)
		assert.Equal(t, test.wantExitCode, exitCode, test.scenario)
	}
}
