package cli

import (
	"flag"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"

	"github.com/sirupsen/logrus"
)

type createIndices struct {
	logger    *logrus.Logger
	shouldRun *bool
	force     *bool
	esClient  elasticsearch.ClientInterface
}

func NewCreateIndices(logger *logrus.Logger) *createIndices {
	esClient, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Fatal(err)
	}

	return &createIndices{
		logger:   logger,
		esClient: esClient,
	}
}

func (c *createIndices) Name() string {
	return "create-indices"
}

func (c *createIndices) DefineFlags() {
	c.shouldRun = flag.Bool("create-indices", false, "create elasticsearch indices")
	c.force = flag.Bool("force", false, "force changes")
}

func (c *createIndices) ShouldRun() bool {
	return *c.shouldRun
}

func (c *createIndices) Run(args []string) error {
	exists, err := c.esClient.IndexExists(person.Person{})
	if err != nil {
		return err
	}

	if exists {
		c.logger.Println("Person index already exists")

		if !*c.force {
			return nil
		}

		c.logger.Println("Changes are forced, deleting old index")

		if err := c.esClient.DeleteIndex(person.Person{}); err != nil {
			return err
		}

		c.logger.Println("Person index deleted successfully")
	}

	if _, err := c.esClient.CreateIndex(person.Person{}); err != nil {
		return err
	}

	c.logger.Println("Person index created successfully")
	return nil
}
