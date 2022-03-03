package cmd

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type CleanupIndicesClient interface {
	ResolveAlias(string) (string, error)
	Indices(string) ([]string, error)
	DeleteIndex(string) error
}

type cleanupCommand struct {
	logger *logrus.Logger
	client CleanupIndicesClient
	index  string
}

type cleanupIndicesCommand struct {
	commands []*cleanupCommand
}

func NewCleanupIndices(logger *logrus.Logger, client CleanupIndicesClient, indexes map[string][]byte) *cleanupIndicesCommand {
	commandArray := &cleanupIndicesCommand{}

	for indexName, _ := range indexes {
		indexCommand := &cleanupCommand{
			logger: logger,
			client: client,
			index:  indexName,
		}
		commandArray.commands = append(commandArray.commands, indexCommand)
	}

	return commandArray
}

func (c *cleanupIndicesCommand) Name() string {
	return "cleanup-indices"
}

func (c *cleanupIndicesCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("cleanup-indices", flag.ExitOnError)

	explain := flagset.Bool("explain", false, "explain the changes that will be made")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for _, s := range c.commands {
		aliasName := strings.Split(s.index, "_")[0]
		aliasedIndex, err := s.client.ResolveAlias(aliasName)
		if err != nil {
			return err
		}
		if aliasedIndex != s.index {
			return fmt.Errorf("alias '%s' does not match current index '%s'", aliasName, s.index)
		}

		indices, err := s.client.Indices(aliasName + "_*")
		if err != nil {
			return err
		}

		for _, index := range indices {
			if index != s.index {
				if *explain {
					s.logger.Println("will delete", index)
				} else {
					if err := s.client.DeleteIndex(index); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
