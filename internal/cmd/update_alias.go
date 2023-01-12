package cmd

import (
	"context"
	"flag"
	"strings"

	"github.com/sirupsen/logrus"
)

type UpdateAliasClient interface {
	ResolveAlias(ctx context.Context, alias string) (string, error)
	UpdateAlias(ctx context.Context, alias, oldIndex, newIndex string) error
}

type updateAliasCommand struct {
	logger         *logrus.Logger
	client         UpdateAliasClient
	currentIndices map[string][]byte
}

func NewUpdateAlias(logger *logrus.Logger, client UpdateAliasClient, indices map[string][]byte) *updateAliasCommand {
	return &updateAliasCommand{
		logger:         logger,
		client:         client,
		currentIndices: indices,
	}
}

func (c *updateAliasCommand) Info() (name, description string) {
	return "update-alias", "update aliases to refer to the current indices"
}

func (c *updateAliasCommand) Run(args []string) error {
	ctx := context.Background()
	flagset := flag.NewFlagSet("update-alias", flag.ExitOnError)

	explain := flagset.Bool("explain", false, "explain the changes that will be made")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for indexName := range c.currentIndices {
		aliasName := strings.Split(indexName, "_")[0]

		aliasedIndex, err := c.client.ResolveAlias(ctx, aliasName)
		if err != nil {
			return err
		}

		if aliasedIndex == indexName {
			c.logger.Printf("alias '%s' is already set to '%s'", aliasName, indexName)
			continue
		}

		if *explain {
			c.logger.Printf("will update alias '%s' to '%s'", aliasName, indexName)
		} else {
			if err := c.client.UpdateAlias(ctx, aliasName, aliasedIndex, indexName); err != nil {
				return err
			}
		}
	}

	return nil
}
