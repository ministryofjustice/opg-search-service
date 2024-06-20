package cmd

import (
	"context"
	"errors"
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
		// assumption is that the alias looks like <alias>_<hash>, where <hash> is appended by opensearch;
		// <alias> may contain underscores, and we assume only the last _<hash> part can be discarded, with the
		// remainder being the alias
		aliasParts := strings.Split(indexName, "_")
		if len(aliasParts) < 2 {
			return errors.New("invalid index alias")
		}
		aliasName := strings.Join(aliasParts[0:len(aliasParts)-1], "_")

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
