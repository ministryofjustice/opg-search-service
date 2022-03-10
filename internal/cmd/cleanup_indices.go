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

type cleanupIndicesCommand struct {
	logger 		*logrus.Logger
	client 		CleanupIndicesClient
	indices		map[string][]byte
}

func NewCleanupIndices(logger *logrus.Logger, client CleanupIndicesClient, indices map[string][]byte) *cleanupIndicesCommand {
	return &cleanupIndicesCommand{
		logger: 	logger,
		client: 	client,
		indices:  	indices,
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

	for indexName := range c.indices {
		aliasName := strings.Split(indexName, "_")[0]
		aliasedIndex, err := c.client.ResolveAlias(aliasName)
		if err != nil {
			return err
		}
		if aliasedIndex != indexName {
			return fmt.Errorf("alias '%s' does not match current index '%s'", aliasName, indexName)
		}

		indices, err := c.client.Indices(aliasName + "_*")
		if err != nil {
			return err
		}

		for _, index := range indices {
			if index != indexName {
				if *explain {
					c.logger.Println("will delete", index)
				} else {
					if err := c.client.DeleteIndex(index); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
