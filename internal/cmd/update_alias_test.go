package cmd

import (
	"testing"

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

func TestUpdateAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_old", nil)

	client.
		On("UpdateAlias", person.AliasName, "person_old", "person_expected").
		Return(nil)

	command := NewUpdateAliasForPersonAndFirm(l, client, "person_expected", "firm_expected")
	assert.Nil(t, command.Run([]string{}))
}

func TestUpdateAliasWhenAliasIsCurrent(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_expected", nil)

	command := NewUpdateAliasForPersonAndFirm(l, client, "person_expected", "firm_expected")
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'person' is already set to 'person_expected'", hook.LastEntry().Message)
}

func TestUpdateAliasSet(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_old", nil)

	client.
		On("UpdateAlias", person.AliasName, "person_old", "person_this").
		Return(nil)

	command := NewUpdateAliasForPersonAndFirm(l, client, "person_unexpected", "firm_expected")
	assert.Nil(t, command.Run([]string{"-set", "person_this"}))
}
