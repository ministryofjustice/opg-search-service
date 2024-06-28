package cmd

import (
	"context"
	"flag"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
)

type IndexClient interface {
	CreateIndex(ctx context.Context, name string, config []byte, force bool) error
	ResolveAlias(ctx context.Context, name string) (string, error)
	CreateAlias(ctx context.Context, alias, index string) error
}

type CreateIndicesCommand struct {
	esClient IndexClient
	indices  []IndexConfig
}

func NewCreateIndices(esClient IndexClient, indices []IndexConfig) *CreateIndicesCommand {
	return &CreateIndicesCommand{
		esClient: esClient,
		indices:  indices,
	}
}

func (c *CreateIndicesCommand) Info() (name, description string) {
	return "create-indices", "create indices"
}

func (c *CreateIndicesCommand) Run(args []string) error {
	ctx := context.Background()
	flagset := flag.NewFlagSet("create-indices", flag.ExitOnError)

	force := flagset.Bool("force", false, "force recreation if index already exists")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for _, indexConfig := range c.indices {
		if err := c.esClient.CreateIndex(ctx, indexConfig.Name, indexConfig.Config, *force); err != nil {
			return err
		}

		_, err := c.esClient.ResolveAlias(ctx, indexConfig.Alias)

		if err == elasticsearch.ErrAliasMissing {
			if err := c.esClient.CreateAlias(ctx, indexConfig.Alias, indexConfig.Name); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return nil
}
