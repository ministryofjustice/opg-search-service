package cli

import (
	"flag"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"os"

	"github.com/sirupsen/logrus"
)

type createIndices struct {
	logger    *logrus.Logger
	shouldRun *bool
	force     *bool
	esClient  elasticsearch.ClientInterface
	exit      func(code int)
}

func NewCreateIndices(logger *logrus.Logger) *createIndices {
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
	c.force = flag.Bool("force", false, "force changes")
}

func (c *createIndices) ShouldRun() bool {
	return *c.shouldRun
}

func (c *createIndices) Run() {
	exists, err := c.esClient.IndexExists(person.Person{})
	if err != nil {
		c.logger.Println(err)
		c.exit(1)
		return
	}

	if exists {
		c.logger.Println("Person index already exists")

		if *c.force {
			c.logger.Println("Changes are forced, deleting old index")
			err := c.esClient.DeleteIndex(person.Person{})

			if err != nil {
				c.logger.Println(err)
				c.exit(1)
				return
			}

			c.logger.Println("Person index deleted successfully")
		} else {
			c.exit(0)
			return
		}
	}

	_, err = c.esClient.CreateIndex(person.Person{})
	if err != nil {
		c.logger.Println(err)
		c.exit(1)
		return
	}
	c.logger.Println("Person index created successfully")
	c.exit(0)
}
