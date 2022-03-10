package cmd

import (
	"flag"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"strings"
)

type IndexClient interface {
	CreateIndex(name string, config []byte, force bool) error
	ResolveAlias(name string) (string, error)
	CreateAlias(alias, index string) error
}

type createIndicesCommand struct {
	esClient    IndexClient
	indexes		map[string][]byte
}

func NewCreateIndices(esClient IndexClient, indexes map[string][]byte) *createIndicesCommand {
	return &createIndicesCommand{
		esClient:   esClient,
		indexes: 	indexes,
	}
}

func (c *createIndicesCommand) Name() string {
	return "create-indices"
}

func (c *createIndicesCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("create-indices", flag.ExitOnError)

	force := flagset.Bool("force", false, "force recreation if index already exists")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for indexName, indexConfig := range c.indexes {
		if err := c.esClient.CreateIndex(indexName, indexConfig, *force); err != nil {
			return err
		}

		typeOfIndex := strings.Split(indexName, "_")[0]
		_, err := c.esClient.ResolveAlias(typeOfIndex)

		if err == elasticsearch.ErrAliasMissing {
			if err := c.esClient.CreateAlias(typeOfIndex, indexName); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
