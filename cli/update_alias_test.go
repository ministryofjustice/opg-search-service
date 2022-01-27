package cli

import (
	"opg-search-service/person"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

type mockUpdateAliasClient struct {
	mock.Mock
}

func (m *mockUpdateAliasClient) UpdateAlias(alias, index string) error {
	args := m.Called(alias, index)
	return args.Error(0)
}

func TestUpdateAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("UpdateAlias", person.AliasName, "person_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, "person_expected")
	command.Run([]string{})
}

func TestUpdateAliasSet(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("UpdateAlias", person.AliasName, "person_this").
		Return(nil)

	command := NewUpdateAlias(l, client, "person_unexpected")
	command.Run([]string{"-set", "person_this"})
}
