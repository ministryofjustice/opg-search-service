package cli

import (
	"flag"

	"github.com/sirupsen/logrus"
)

type Command interface {
	Name() string
	DefineFlags()
	ShouldRun() bool
	Run(args []string) error
}

func Run(logger *logrus.Logger, cmds ...Command) {
	for _, cmd := range cmds {
		cmd.DefineFlags()
	}

	flag.Parse()
	args := flag.Args()

	if len(args) > 0 {
		for _, cmd := range cmds {
			if cmd.Name() == args[0] {
				logger.Printf("Running command: %T", cmd)
				if err := cmd.Run(args[1:]); err != nil {
					logger.Errorln(err)
					logger.Exit(1)
					return
				}

				logger.Exit(0)
			}
		}
	} else {
		for _, cmd := range cmds {
			if cmd.ShouldRun() {
				logger.Printf("Running command: %T", cmd)
				if err := cmd.Run(args); err != nil {
					logger.Errorln(err)
					logger.Exit(1)
					return
				}

				logger.Exit(0)
			}
		}
	}
}
