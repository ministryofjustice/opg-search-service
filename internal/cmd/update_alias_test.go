package cmd

import (
	"github.com/ministryofjustice/opg-search-service/internal/deputy"
	"testing"

	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUpdateAliasClient struct {
	mock.Mock
}

func (m *mockUpdateAliasClient) ResolveAlias(alias string) (string, error) {
	args := m.Called(alias)
	return args.String(0), args.Error(1)
}

func (m *mockUpdateAliasClient) UpdateAlias(alias, oldIndex, newIndex string) error {
	args := m.Called(alias, oldIndex, newIndex)
	return args.Error(0)
}

func TestUpdatePersonAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_old", nil)

	client.
		On("UpdateAlias", person.AliasName, "person_old", "person_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"person_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))
}

func TestUpdateFirmAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", firm.AliasName).
		Return("firm_old", nil)

	client.
		On("UpdateAlias", firm.AliasName, "firm_old", "firm_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))
}

func TestUpdateDeputyAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", deputy.AliasName).
		Return("deputy_old", nil)

	client.
		On("UpdateAlias", deputy.AliasName, "deputy_old", "deputy_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"deputy_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))
}

func TestUpdatePersonAliasWhenAliasIsCurrent(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_expected", nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"person_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'person' is already set to 'person_expected'", hook.LastEntry().Message)
}

func TestUpdateFirmAliasWhenAliasIsCurrent(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", firm.AliasName).
		Return("firm_expected", nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'firm' is already set to 'firm_expected'", hook.LastEntry().Message)
}

func TestUpdateDeputyAliasWhenAliasIsCurrent(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", deputy.AliasName).
		Return("deputy_expected", nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"deputy_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'deputy' is already set to 'deputy_expected'", hook.LastEntry().Message)
}
