package cmd

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCleanupIndicesClient struct {
	mock.Mock
}

func (m *mockCleanupIndicesClient) ResolveAlias(ctx context.Context, alias string) (string, error) {
	args := m.Called(ctx, alias)
	return args.String(0), args.Error(1)
}

func (m *mockCleanupIndicesClient) Indices(ctx context.Context, term string) ([]string, error) {
	args := m.Called(ctx, term)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockCleanupIndicesClient) DeleteIndex(ctx context.Context, index string) error {
	args := m.Called(ctx, index)
	return args.Error(0)
}

func TestCleanupIndices(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockCleanupIndicesClient{}

	client.
		On("ResolveAlias", mock.Anything, person.AliasName).
		Return("person_something", nil)

	client.
		On("ResolveAlias", mock.Anything, firm.AliasName).
		Return("firm_something", nil)

	client.
		On("Indices", mock.Anything, "person_*").
		Return([]string{"person_xyz", "person_something", "person_abc"}, nil)

	client.
		On("Indices", mock.Anything, "firm_*").
		Return([]string{"firm_xyz", "firm_something", "firm_abc"}, nil)

	client.On("DeleteIndex", mock.Anything, "person_xyz").Return(nil).Once()
	client.On("DeleteIndex", mock.Anything, "person_abc").Return(nil).Once()

	client.On("DeleteIndex", mock.Anything, "firm_xyz").Return(nil).Once()
	client.On("DeleteIndex", mock.Anything, "firm_abc").Return(nil).Once()

	command := NewCleanupIndices(
		l,
		client,
		[]IndexConfig{
			{
				Name:   "person_something",
				Alias:  "person",
				Config: indexConfig,
			},
			{
				Name:   "firm_something",
				Alias:  "firm",
				Config: indexConfig,
			},
		},
	)
	assert.Nil(t, command.Run([]string{}))
}

func TestCleanupIndicesExplain(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockCleanupIndicesClient{}

	client.
		On("ResolveAlias", mock.Anything, person.AliasName).
		Return("person_something", nil)

	client.
		On("ResolveAlias", mock.Anything, firm.AliasName).
		Return("firm_something", nil)

	client.
		On("Indices", mock.Anything, "person_*").
		Return([]string{"person_xyz", "person_something", "person_abc"}, nil)

	client.
		On("Indices", mock.Anything, "firm_*").
		Return([]string{"firm_xyz", "firm_something", "firm_abc"}, nil)

	command := NewCleanupIndices(
		l,
		client,
		[]IndexConfig{
			{
				Name:   "person_something",
				Alias:  "person",
				Config: indexConfig,
			},
			{
				Name:   "firm_something",
				Alias:  "firm",
				Config: indexConfig,
			},
		},
	)

	assert.Nil(t, command.Run([]string{"-explain"}))

	expected := []string{
		"will delete person_xyz",
		"keeping index person_something aliased as person",
		"will delete person_abc",
		"will delete firm_xyz",
		"keeping index firm_something aliased as firm",
		"will delete firm_abc",
	}
	if assert.Len(t, hook.Entries, len(expected)) {
		for i, e := range hook.Entries {
			assert.Equal(t, expected[i], e.Message)
		}
	}
}
