package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

type Command interface {
	Name() string
	Run(args []string) error
}

func Run(logger *logrus.Logger, cmds ...Command) {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), `Usage of %s <command> [arguments]:

Commands:
	hc                		run healthcheck on the search service
	create-indices    		create elasticsearch indices
	index             		index person records

	create-indices person	create person index
	create-indices firm		create firm index
	index             		index all records
	index person			index person records
	index firm				index firm records

`, os.Args[0])
	}
	flag.Parse()
	args := flag.Args()

	if len(args) > 0 {
		for _, cmd := range cmds {
			if cmd.Name() == args[0] {
				logger.Printf("Running command: %T", cmd)
				//diff
				if (len(args)>1){
					logger.Println(args[1])
					logger.Println(args[1:])
				}
				err := cmd.Run(args[1:])
				if err != nil {
					logger.Errorln(err)
					logger.Exit(1)
					return
				}

				logger.Exit(0)
			}
		}
	}
}
