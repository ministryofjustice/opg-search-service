package cli

import (
	"flag"
)

type IndexClient interface {
	CreateIndex(name string, config []byte, force bool) error
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

	return nil
}
