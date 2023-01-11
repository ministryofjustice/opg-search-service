package cmd

import (
	"context"
	"flag"
	"strings"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
)

type IndexClient interface {
	CreateIndex(ctx context.Context, name string, config []byte, force bool) error
	ResolveAlias(ctx context.Context, name string) (string, error)
	CreateAlias(ctx context.Context, alias, index string) error
}

type createIndicesCommand struct {
	esClient IndexClient
	indices  map[string][]byte
}

func NewCreateIndices(esClient IndexClient, indices map[string][]byte) *createIndicesCommand {
	return &createIndicesCommand{
		esClient: esClient,
		indices:  indices,
	}
}

func (c *createIndicesCommand) Info() (name, description string) {
	return "create-indices", "create indices"
}

func (c *createIndicesCommand) Run(args []string) error {
	ctx := context.Background()
	flagset := flag.NewFlagSet("create-indices", flag.ExitOnError)

	force := flagset.Bool("force", false, "force recreation if index already exists")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for indexName, indexConfig := range c.indices {
		if err := c.esClient.CreateIndex(ctx, indexName, indexConfig, *force); err != nil {
			return err
		}

		aliasName := strings.Split(indexName, "_")[0]
		_, err := c.esClient.ResolveAlias(ctx, aliasName)

		if err == elasticsearch.ErrAliasMissing {
			if err := c.esClient.CreateAlias(ctx, aliasName, indexName); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
