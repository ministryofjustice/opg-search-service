package cli

import (
	"flag"
	"log"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"os"
)

type createIndices struct {
	logger    *log.Logger
	shouldRun *bool
	esClient  elasticsearch.ClientInterface
	exit      func(code int)
}

func NewCreateIndices(logger *log.Logger) *createIndices {
	esClient, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Fatal(err)
	}

	return &createIndices{
		logger:   logger,
		esClient: esClient,
		exit:     os.Exit,
	}
}

func (c *createIndices) DefineFlags() {
	c.shouldRun = flag.Bool("create-indices", false, "create elasticsearch indices")
}

func (c *createIndices) ShouldRun() bool {
	return *c.shouldRun
}

func (c *createIndices) Run() {
	_, err := c.esClient.CreateIndex(person.Person{})
	if err != nil {
		c.logger.Println(err)
		c.exit(1)
		return
	}
	c.logger.Println("Person index created successfully")
	c.exit(0)
}