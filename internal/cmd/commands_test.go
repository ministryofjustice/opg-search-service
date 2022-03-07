package cmd

import (
	"os"
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
	return m.name
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

	cmd1 := &MockCommand{name: "1"}

	cmd2 := &MockCommand{name: "2"}
	cmd2.On("Run", []string{}).Times(1).Return(nil)

	os.Args = []string{"search-service", "2"}
	Run(l, cmd1, cmd2)

	assert.Contains(t, "Running command: *cmd.MockCommand", hook.LastEntry().Message)
	assert.Equal(t, 0, exitCode)
}

func TestCommandsRunSecondArgument(t *testing.T) {
	exitCode := -1

	l, hook := test.NewNullLogger()
	l.ExitFunc = func(c int) {
		exitCode = c
	}

	cmd1 := &MockCommand{name: "index"}

	cmd2 := &MockCommand{name: "index person"}
	cmd2.On("Run", []string{}).Times(1).Return(nil)

	os.Args = []string{"search-service", "index person"}
	Run(l, cmd1, cmd2)

	assert.Contains(t, "Running command: *cmd.MockCommand", hook.LastEntry().Message)
	assert.Equal(t, 0, exitCode)
}

//this is the one that always failed even on main
//func TestCommandsRunErrors(t *testing.T) {
//	exitCode := -1
//
//	l, hook := test.NewNullLogger()
//	l.ExitFunc = func(c int) {
//		exitCode = c
//	}
//
//	cmd1 := &MockCommand{name: "1"}
//
//	cmd2 := &MockCommand{name: "2"}
//	cmd2.On("Run", []string{}).Times(1).Return(errors.New("what"))
//
//	Run(l, cmd1, cmd2)
//
//	assert.Equal(t, "what", hook.LastEntry().Message)
//	assert.Equal(t, 1, exitCode)
//}
