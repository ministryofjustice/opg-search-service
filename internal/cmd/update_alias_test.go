package cmd

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/digitallpa"
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

func (m *mockUpdateAliasClient) ResolveAlias(ctx context.Context, alias string) (string, error) {
	args := m.Called(ctx, alias)
	return args.String(0), args.Error(1)
}

func (m *mockUpdateAliasClient) UpdateAlias(ctx context.Context, alias, oldIndex, newIndex string) error {
	args := m.Called(ctx, alias, oldIndex, newIndex)
	return args.Error(0)
}

func TestUpdatePersonAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", mock.Anything, person.AliasName).
		Return("person_old", nil)

	client.
		On("UpdateAlias", mock.Anything, person.AliasName, "person_old", "person_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"person_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))
}

func TestUpdateFirmAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", mock.Anything, firm.AliasName).
		Return("firm_old", nil)

	client.
		On("UpdateAlias", mock.Anything, firm.AliasName, "firm_old", "firm_expected").
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))
}

func TestUpdatePersonAliasWhenAliasIsCurrent(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", mock.Anything, person.AliasName).
		Return("person_expected", nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"person_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'person' is already set to 'person_expected'", hook.LastEntry().Message)
}

func TestUpdateFirmAliasWhenAliasIsCurrent(t *testing.T) {
	l, hook := test.NewNullLogger()
	client := &mockUpdateAliasClient{}

	client.
		On("ResolveAlias", mock.Anything, firm.AliasName).
		Return("firm_expected", nil)

	command := NewUpdateAlias(l, client, map[string][]byte{"firm_expected": indexConfig})
	assert.Nil(t, command.Run([]string{}))

	assert.Equal(t, "alias 'firm' is already set to 'firm_expected'", hook.LastEntry().Message)
}

func TestUpdateDigitalLpaAlias(t *testing.T) {
	l, _ := test.NewNullLogger()
	client := &mockUpdateAliasClient{}
	oldAlias := fmt.Sprintf("%s_1a2b3c4d", digitallpa.AliasName)
	newAlias := fmt.Sprintf("%s_0z9y8x7w", digitallpa.AliasName)

	client.
		On("ResolveAlias", mock.Anything, digitallpa.AliasName).
		Return(oldAlias, nil)

	client.
		On("UpdateAlias", mock.Anything, digitallpa.AliasName, oldAlias, newAlias).
		Return(nil)

	command := NewUpdateAlias(l, client, map[string][]byte{newAlias: indexConfig})
	assert.Nil(t, command.Run([]string{}))
}
