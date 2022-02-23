package cmd

import (
	"flag"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"strings"

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

	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.Println("in update alias")
	l.Println(c.index)
	indexName := strings.Split(c.index, "_")[0]

	set := flagset.String("set", c.index, "index to point the alias at")

	if err := flagset.Parse(args); err != nil {
		return err
	}
	var aliasName string
	if indexName == person.AliasName {
		aliasName = person.AliasName
	} else {
		aliasName = firm.AliasName
	}
	aliasedIndex, err := c.client.ResolveAlias(aliasName)
	if err != nil {
		return err
	}

	if aliasedIndex == *set {
		c.logger.Printf("alias '%s' is already set to '%s'", aliasName, *set)
		return nil
	}

	return c.client.UpdateAlias(aliasName, aliasedIndex, *set)
}
