package cmd

import (
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCleanupIndicesClient struct {
	mock.Mock
}

func (m *mockCleanupIndicesClient) ResolveAlias(alias string) (string, error) {
	args := m.Called(alias)
	return args.String(0), args.Error(1)
}

func (m *mockCleanupIndicesClient) Indices(term string) ([]string, error) {
	args := m.Called(term)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockCleanupIndicesClient) DeleteIndex(index string) error {
	args := m.Called(index)
	return args.Error(0)
}

func TestCleanupIndicesForPersonAndFirm2(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockCleanupIndicesClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_something", nil)

	client.
		On("Indices", "person_*").
		Return([]string{"person_xyz", "person_something", "person_abc"}, nil)

	client.
		On("ResolveAlias", firm.AliasName).
		Return("firm_something", nil)

	client.
		On("Indices", "firm_*").
		Return([]string{"firm_xyz", "firm_something", "firm_abc"}, nil)

	client.On("DeleteIndex", "person_xyz").Return(nil).Once()
	client.On("DeleteIndex", "person_abc").Return(nil).Once()

	client.On("DeleteIndex", "firm_xyz").Return(nil).Once()
	client.On("DeleteIndex", "firm_abc").Return(nil).Once()

	command := NewCleanupIndicesForPersonAndFirm(l, client, "person_something", "firm_something")
	assert.Nil(t, command.Run([]string{}))
}

func TestCleanupIndicesWhenAliasNotCurrent(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockCleanupIndicesClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_xyz", nil)

	client.
		On("Indices", "person_*").
		Return([]string{"person_xyz", "person_something", "person_abc"}, nil)

	command := NewCleanupIndicesForPersonAndFirm(l, client, "person_something", "firm_something")
	assert.NotNil(t, command.Run([]string{}))
}

func TestCleanupIndicesExplain(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockCleanupIndicesClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_something", nil)

	client.
		On("Indices", "person_*").
		Return([]string{"person_xyz", "person_something", "person_abc"}, nil)

	client.
		On("ResolveAlias", firm.AliasName).
		Return("firm_something", nil)

	client.
		On("Indices", "firm_*").
		Return([]string{"firm_xyz", "firm_something", "firm_abc"}, nil)

	command := NewCleanupIndicesForPersonAndFirm(l, client, "person_something","firm_something")
	assert.Nil(t, command.Run([]string{"-explain"}))

	expected := []string{"will delete person_xyz", "will delete person_abc", "will delete firm_xyz", "will delete firm_abc"}
	if assert.Len(t, hook.Entries, len(expected)) {
		for i, e := range hook.Entries {
			assert.Equal(t, expected[i], e.Message)
		}
	}
}
