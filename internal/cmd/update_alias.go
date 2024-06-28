package cmd

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
)

type UpdateAliasClient interface {
	ResolveAlias(ctx context.Context, alias string) (string, error)
	UpdateAlias(ctx context.Context, alias, oldIndex, newIndex string) error
}

type UpdateAliasCommand struct {
	logger         *logrus.Logger
	client         UpdateAliasClient
	currentIndices []IndexConfig
}

func NewUpdateAlias(logger *logrus.Logger, client UpdateAliasClient, currentIndices []IndexConfig) *UpdateAliasCommand {
	return &UpdateAliasCommand{
		logger:         logger,
		client:         client,
		currentIndices: currentIndices,
	}
}

func (c *UpdateAliasCommand) Info() (name, description string) {
	return "update-alias", "update aliases to refer to the current indices"
}

func (c *UpdateAliasCommand) Run(args []string) error {
	ctx := context.Background()
	flagset := flag.NewFlagSet("update-alias", flag.ExitOnError)

	explain := flagset.Bool("explain", false, "explain the changes that will be made")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for _, indexConfig := range c.currentIndices {
		currentAliasedIndex, err := c.client.ResolveAlias(ctx, indexConfig.Alias)
		if err != nil {
			return err
		}

		if currentAliasedIndex == indexConfig.Name {
			c.logger.Printf("alias '%s' is already set to '%s'", indexConfig.Alias, indexConfig.Name)
			continue
		}

		if *explain {
			c.logger.Printf("will update alias '%s' to '%s'", indexConfig.Alias, indexConfig.Name)
		} else {
			if err := c.client.UpdateAlias(ctx, indexConfig.Alias, currentAliasedIndex, indexConfig.Name); err != nil {
				return err
			}
		}
	}

	return nil
}
