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

type indicesCommand struct {
	esClient  IndexClient
	indexName string
	indexConfig []byte
}

type createIndicesCommand struct {
	commands []*indicesCommand
}

func NewCreateIndices(esClient IndexClient, indexes map[string][]byte) *createIndicesCommand {
	commandArray := &createIndicesCommand{}

	for indexName, config := range indexes {
		indicesCommand := &indicesCommand{
			esClient:    esClient,
			indexName:   indexName,
			indexConfig: config,
		}
		commandArray.commands = append(commandArray.commands, indicesCommand)
	}

	return commandArray
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

	for _, s := range c.commands {
		if err := s.esClient.CreateIndex(s.indexName, s.indexConfig, *force); err != nil {
			return err
		}

		typeOfIndex := strings.Split(s.indexName, "_")[0]
		_, err := s.esClient.ResolveAlias(typeOfIndex)

		if err == elasticsearch.ErrAliasMissing {
			if err := s.esClient.CreateAlias(typeOfIndex, s.indexName); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

	}

	return nil
}
