package cli

import (
	"flag"
	"fmt"
	"opg-search-service/person"

	"github.com/sirupsen/logrus"
)

type CleanupIndicesClient interface {
	ResolveAlias(string) (string, error)
	Indices(string) ([]string, error)
	DeleteIndex(string) error
}

type cleanupIndicesCommand struct {
	logger *logrus.Logger
	client CleanupIndicesClient
	index  string
}

func NewCleanupIndices(logger *logrus.Logger, client CleanupIndicesClient, index string) *cleanupIndicesCommand {
	return &cleanupIndicesCommand{
		logger: logger,
		client: client,
		index:  index,
	}
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

	aliasedIndex, err := c.client.ResolveAlias(person.AliasName)
	if err != nil {
		return err
	}
	if aliasedIndex != c.index {
		return fmt.Errorf("alias '%s' does not match current index '%s'", person.AliasName, c.index)
	}

	indices, err := c.client.Indices("person_*")
	if err != nil {
		return err
	}

	for _, index := range indices {
		if index != c.index {
			if *explain {
				c.logger.Println("will delete", index)
			} else {
				if err := c.client.DeleteIndex(index); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
