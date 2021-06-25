package cli

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) DefineFlags() {
	_ = m.Called()
}

func (m *MockCommand) ShouldRun() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockCommand) Run() {
	_ = m.Called()
}

func TestCommands(t *testing.T) {
	l, _ := test.NewNullLogger()
	c := Commands(l)

	assert.IsType(t, new(commands), c)
	assert.Same(t, l, c.logger)
}

func TestCommands_Register(t *testing.T) {
	l, hook := test.NewNullLogger()

	cmd1 := new(MockCommand)
	cmd2 := new(MockCommand)

	cmd1.On("DefineFlags").Times(1).Return()
	cmd2.On("DefineFlags").Times(1).Return()

	cmd1.On("ShouldRun").Times(1).Return(false)
	cmd2.On("ShouldRun").Times(1).Return(true)

	cmd2.On("Run").Times(1).Return()

	c := commands{logger: l}
	c.Register(cmd1, cmd2)

	assert.Contains(t, hook.LastEntry().Message, "Running command: *cli.MockCommand")
}
