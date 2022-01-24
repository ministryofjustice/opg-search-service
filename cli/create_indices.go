package cli

import (
	"flag"
	"net/http"
	"opg-search-service/elasticsearch"

	"github.com/sirupsen/logrus"
)

type IndexClient interface {
	CreateIndex(name string, config []byte, force bool) error
}

type createIndices struct {
	esClient    IndexClient
	indexName   string
	indexConfig []byte
}

func NewCreateIndices(logger *logrus.Logger, indexName string, indexConfig []byte) *createIndices {
	esClient, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Fatal(err)
	}

	return &createIndices{
		esClient:    esClient,
		indexName:   indexName,
		indexConfig: indexConfig,
	}
}

func (c *createIndices) Name() string {
	return "create-indices"
}

func (c *createIndices) Run(args []string) error {
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
