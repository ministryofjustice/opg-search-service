package cmd

import (
	"flag"
	"github.com/sirupsen/logrus"
)

type IndexClient interface {
	CreateIndex(name string, config []byte, force bool) error
}

type createIndicesCommand struct {
	esClient    		IndexClient
	indexName		   	string
	indexConfig		 	[]byte
}

type createIndicesCommandForPersonAndFirm struct {
	esClient    		IndexClient
	personIndexName   	string
	personIndexConfig 	[]byte
	firmIndexName		string
	firmIndexConfig		[]byte
}

func NewCreateIndices(esClient IndexClient, indexName string, indexConfig []byte) *createIndicesCommand {
	return &createIndicesCommand{
		esClient:    esClient,
		indexName:   indexName,
		indexConfig: indexConfig,
	}
}

func NewCreateIndicesForPersonAndFirm(esClient IndexClient, personIndexName string, personIndexConfig []byte, firmIndexName string, firmIndexConfig []byte) *createIndicesCommandForPersonAndFirm {
	return &createIndicesCommandForPersonAndFirm{
		esClient:    		esClient,
		personIndexName:   	personIndexName,
		personIndexConfig: 	personIndexConfig,
		firmIndexName:		firmIndexName,
		firmIndexConfig:	firmIndexConfig,
	}
}

//func (c *createIndicesCommand) Name() string {
//	return "create-indices"
//}

func (c *createIndicesCommandForPersonAndFirm) Name() string {
	return "create-indices"
}

//func (c *createIndicesCommand) Run(args []string) error {
//
//	l := logrus.New()
//	l.SetFormatter(&logrus.JSONFormatter{})
//
//	l.Println("Running indices")
//	l.Println(args)
//	l.Println(c.indexName)
//
//	flagset := flag.NewFlagSet("create-indices", flag.ExitOnError)
//
//	force := flagset.Bool("force", false, "force recreation if index already exists")
//
//	if err := flagset.Parse(args); err != nil {
//		return err
//	}
//
//	if err := c.esClient.CreateIndex(c.indexName, c.indexConfig, *force); err != nil {
//		return err
//	}
//
//	return nil
//}

func (c *createIndicesCommandForPersonAndFirm) Run(args []string) error {

	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})

	l.Println("Running indices")
	l.Println(c.personIndexName)
	l.Println(c.firmIndexName)


	flagset := flag.NewFlagSet("create-indices", flag.ExitOnError)

	force := flagset.Bool("force", false, "force recreation if index already exists")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	if err := c.esClient.CreateIndex(c.personIndexName, c.personIndexConfig, *force); err != nil {
		return err
	}

	if err := c.esClient.CreateIndex(c.firmIndexName, c.firmIndexConfig, *force); err != nil {
		return err
	}

	return nil
}
