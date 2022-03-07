package cmd

import (
	"github.com/ministryofjustice/opg-search-service/internal/indices"
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
//check because the flag is panicking that its being redefined
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
		On("ResolveAlias", indices.AliasNameFirm).
		Return("firm_old", nil)

	client.
		On("UpdateAlias", indices.AliasNameFirm, "firm_old", "firm_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
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
		On("ResolveAlias", indices.AliasNameFirm).
		Return("firm_expected", nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'firm' is already set to 'firm_expected'", hook.LastEntry().Message)
}

func TestUpdatePersonAliasSet(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", person.AliasName).
		Return("person_old", nil)

	client.
		On("UpdateAlias", person.AliasName, "person_old", "person_this").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"person_expected": indexConfig})
	assert.Nil(t, command.Run([]string{"-set", "person_this"}))
}

func TestUpdateFirmAliasSet(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", indices.AliasNameFirm).
		Return("firm_old", nil)

	client.
		On("UpdateAlias", indices.AliasNameFirm, "firm_old", "firm_this").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
	assert.Nil(t, command.Run([]string{"-set", "firm_this"}))
}