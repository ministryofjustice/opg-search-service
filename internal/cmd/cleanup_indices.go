package cmd

import (
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
)

type CleanupIndicesClient interface {
	ResolveAlias(ctx context.Context, alias string) (string, error)
	Indices(ctx context.Context, term string) ([]string, error)
	DeleteIndex(ctx context.Context, name string) error
}

type CleanupIndicesCommand struct {
	logger         *logrus.Logger
	client         CleanupIndicesClient
	currentIndices []IndexConfig
}

func NewCleanupIndices(logger *logrus.Logger, client CleanupIndicesClient, currentIndices []IndexConfig) *CleanupIndicesCommand {
	return &CleanupIndicesCommand{
		logger:         logger,
		client:         client,
		currentIndices: currentIndices,
	}
}

func (c *CleanupIndicesCommand) Info() (name, description string) {
	return "cleanup-indices", "remove unused indices"
}

func (c *CleanupIndicesCommand) Run(args []string) error {
	ctx := context.Background()
	flagset := flag.NewFlagSet("cleanup-indices", flag.ExitOnError)

	explain := flagset.Bool("explain", false, "explain the changes that will be made")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for _, indexConfig := range c.currentIndices {
		indices, err := c.client.Indices(ctx, indexConfig.Alias+"_*")
		if err != nil {
			return err
		}

		for _, indexName := range indices {
			if indexConfig.Name == indexName {
				c.logger.Println(fmt.Sprintf("keeping index %s aliased as %s", indexName, indexConfig.Alias))
				continue
			}

			if *explain {
				c.logger.Println("will delete", indexName)
			} else {
				if err := c.client.DeleteIndex(ctx, indexName); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
