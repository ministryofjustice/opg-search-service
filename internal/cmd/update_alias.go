package cmd

import (
	"flag"

	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus"
)

type UpdateAliasClient interface {
	ResolveAlias(string) (string, error)
	UpdateAlias(string, string, string) error
}

type updateAliasCommand struct {
	logger *logrus.Logger
	client UpdateAliasClient
	index  string
}

func NewUpdateAlias(logger *logrus.Logger, client UpdateAliasClient, index string) *updateAliasCommand {
	return &updateAliasCommand{
		logger: logger,
		client: client,
		index:  index,
	}
}

func (c *updateAliasCommand) Name() string {
	return "update-alias"
}

func (c *updateAliasCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("update-alias", flag.ExitOnError)

	set := flagset.String("set", c.index, "index to point the alias at")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	aliasedIndex, err := c.client.ResolveAlias(person.AliasName)
	if err != nil {
		return err
	}

	if aliasedIndex == *set {
		c.logger.Printf("alias '%s' is already set to '%s'", person.AliasName, *set)
		return nil
	}

	return c.client.UpdateAlias(person.AliasName, aliasedIndex, *set)
}
