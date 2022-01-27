package cli

import (
	"flag"
	"opg-search-service/person"

	"github.com/sirupsen/logrus"
)

type UpdateAliasClient interface {
	UpdateAlias(string, string) error
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

	return c.client.UpdateAlias(person.AliasName, *set)
}
