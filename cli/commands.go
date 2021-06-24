package cli

import (
	"flag"

	"github.com/sirupsen/logrus"
)

type Command interface {
	DefineFlags()
	ShouldRun() bool
	Run()
}

type commands struct {
	logger *logrus.Logger
}

func Commands(logger *logrus.Logger) *commands {
	c := commands{
		logger,
	}
	return &c
}

func (c commands) Register(cmds ...Command) {
	for _, cmd := range cmds {
		cmd.DefineFlags()
	}

	flag.Parse()

	for _, cmd := range cmds {
		if cmd.ShouldRun() {
			c.logger.Printf("Running command: %T", cmd)
			cmd.Run()
		}
	}
}
