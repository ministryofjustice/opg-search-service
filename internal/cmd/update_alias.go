package cmd

import (
	"flag"
	"strings"

	"github.com/sirupsen/logrus"
)

type UpdateAliasClient interface {
	ResolveAlias(string) (string, error)
	UpdateAlias(string, string, string) error
}

type aliasCommand struct {
	logger *logrus.Logger
	client UpdateAliasClient
	index  string
}

type updateAliasCommand struct {
	commands []*aliasCommand
}

func NewUpdateAlias(logger *logrus.Logger, client UpdateAliasClient, indexes map[string][]byte) *updateAliasCommand {
	commandArray := &updateAliasCommand{}

	for indexName, _ := range indexes {
		indexCommand := &aliasCommand{
			logger: logger,
			client: client,
			index:  indexName,
		}
		commandArray.commands = append(commandArray.commands, indexCommand)
	}

	return commandArray
}

func (c *updateAliasCommand) Name() string {
	return "update-alias"
}

func (c *updateAliasCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("update-alias", flag.ExitOnError)

	for _, s := range c.commands {
		set := flagset.String("set", s.index, "index to point the alias at")

		if err := flagset.Parse(args); err != nil {
			return err
		}
		aliasName := strings.Split(s.index, "_")[0]
		aliasedIndex, err := s.client.ResolveAlias(aliasName)
		if err != nil {
			return err
		}

		if aliasedIndex == *set {
			s.logger.Printf("alias '%s' is alrFeady set to '%s'", aliasName, *set)
			return nil
		}

		err = s.client.UpdateAlias(aliasName, aliasedIndex, *set)
		if err != nil {
			return err
		}

	}
	return nil
}
