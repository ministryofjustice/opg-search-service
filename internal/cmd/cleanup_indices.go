package cmd

import (
	"flag"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"strings"

	"github.com/ministryofjustice/opg-search-service/internal/person"
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

	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	l.Println("Running cleanup indices")
	l.Println(c.index)

	flagset := flag.NewFlagSet("cleanup-indices", flag.ExitOnError)

	explain := flagset.Bool("explain", false, "explain the changes that will be made")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	indexName := strings.Split(c.index, "_")[0]
	var aliasName string
	if indexName == person.AliasName {
		aliasName = person.AliasName
	} else {
		aliasName = firm.AliasName
	}

	l.Println(aliasName)

	aliasedIndex, err := c.client.ResolveAlias(aliasName)
	if err != nil {
		return err
	}
	if aliasedIndex != c.index {
		return fmt.Errorf("alias '%s' does not match current index '%s'", aliasName, c.index)
	}

	indices, err := c.client.Indices(aliasName + "_*")
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
