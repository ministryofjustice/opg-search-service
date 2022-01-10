package cli

import (
	"errors"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCommand struct {
	name string
	mock.Mock
}

func (m *MockCommand) Name() string {
	return "mock"
}

func (m *MockCommand) DefineFlags() {
	_ = m.Called()
}

func (m *MockCommand) ShouldRun() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockCommand) Run(xs []string) error {
	args := m.Called(xs)
	return args.Error(0)
}

func TestCommandsRun(t *testing.T) {
	exitCode := -1

	l, hook := test.NewNullLogger()
	l.ExitFunc = func(c int) {
		exitCode = c
	}

	cmd1 := &MockCommand{}
	cmd1.On("DefineFlags").Times(1).Return()
	cmd1.On("ShouldRun").Times(1).Return(false)

	cmd2 := &MockCommand{}
	cmd2.On("DefineFlags").Times(1).Return()
	cmd2.On("ShouldRun").Times(1).Return(true)
	cmd2.On("Run", []string{}).Times(1).Return(nil)

	Run(l, cmd1, cmd2)

	assert.Contains(t, "Running command: *cli.MockCommand", hook.LastEntry().Message)
	assert.Equal(t, 0, exitCode)
}

func TestCommandsRunErrors(t *testing.T) {
	exitCode := -1

	l, hook := test.NewNullLogger()
	l.ExitFunc = func(c int) {
		exitCode = c
	}

	cmd1 := &MockCommand{name: "1"}
	cmd1.On("DefineFlags").Times(1).Return()
	cmd1.On("ShouldRun").Times(1).Return(false)

	cmd2 := &MockCommand{name: "2"}
	cmd2.On("DefineFlags").Times(1).Return()
	cmd2.On("ShouldRun").Times(1).Return(true)
	cmd2.On("Run", []string{}).Times(1).Return(errors.New("what"))

	Run(l, cmd1, cmd2)

	assert.Equal(t, "what", hook.LastEntry().Message)
	assert.Equal(t, 1, exitCode)
}
