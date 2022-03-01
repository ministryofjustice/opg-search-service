package cmd

import (
	"flag"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/person"
)

type IndexClient interface {
	CreateIndex(name string, config []byte, force bool) error
	ResolveAlias(name string) (string, error)
	CreateAlias(alias, index string) error
}

type createIndicesCommand struct {
	esClient    IndexClient
	indexName   string
	indexConfig []byte
}

func NewCreateIndices(esClient IndexClient, indexName string, indexConfig []byte) *createIndicesCommand {
	return &createIndicesCommand{
		esClient:    esClient,
		indexName:   indexName,
		indexConfig: indexConfig,
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

	if err := c.esClient.CreateIndex(c.indexName, c.indexConfig, *force); err != nil {
		return err
	}

	_, err := c.esClient.ResolveAlias(person.AliasName)
	if err == elasticsearch.ErrAliasMissing {
		if err := c.esClient.CreateAlias(person.AliasName, c.indexName); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
