package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

type Command interface {
	Info() (name, description string)
	Run(args []string) error
}

func Run(logger *logrus.Logger, cmds ...Command) {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s <command> [arguments]:\n\n Commands:\n", os.Args[0])

		for _, cmd := range cmds {
			name, description := cmd.Info()
			_, _ = fmt.Fprintf(flag.CommandLine.Output(), "  %s %s %s\n", name, strings.Repeat(" ", 20-len(name)), description)
		}
	}
	flag.Parse()
	args := flag.Args()

	if len(args) > 0 {
		for _, cmd := range cmds {
			if name, _ := cmd.Info(); name == args[0] {
				logger.Printf("Running command: %T", cmd)
				if err := cmd.Run(args[1:]); err != nil {
					logger.Errorln(err)
					logger.Exit(1)
					return
				}

				logger.Exit(0)
			}
		}
	}
}
