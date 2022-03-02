package cmd

import (
	"flag"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"strings"

	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus"
)

type UpdateAliasClient interface {
	ResolveAlias(string) (string, error)
	UpdateAlias(string, string, string) error
}

//type updateAliasCommand struct {
//	logger *logrus.Logger
//	client UpdateAliasClient
//	index  string
//}

type updateAliasCommandForPersonAndFirm struct {
	logger 		 	*logrus.Logger
	client 		 	UpdateAliasClient
	personIndex  	string
	firmIndex		string
}

//func NewUpdateAlias(logger *logrus.Logger, client UpdateAliasClient, index string) *updateAliasCommand {
//	return &updateAliasCommand{
//		logger: logger,
//		client: client,
//		index:  index,
//	}
//}

func NewUpdateAliasForPersonAndFirm(logger *logrus.Logger, client UpdateAliasClient, personIndex string, firmIndex string) *updateAliasCommandForPersonAndFirm {
	return &updateAliasCommandForPersonAndFirm{
		logger: 	 	logger,
		client: 		client,
		personIndex:  	personIndex,
		firmIndex:  	firmIndex,
	}
}

//func (c *updateAliasCommand) Name() string {
//	return "update-alias"
//}

func (c *updateAliasCommandForPersonAndFirm) Name() string {
	return "update-alias"
}

func (c *updateAliasCommandForPersonAndFirm) Run(args []string) error {

	flagset := flag.NewFlagSet("update-alias", flag.ExitOnError)

	var indexes = []string{c.personIndex, c.firmIndex}
	var indexesToBeAliased []string
	var aliasName string
	var set *string

	for _, currentIndex := range indexes {
		indexName := strings.Split(currentIndex, "_")[0]
		set = flagset.String("set", currentIndex, "index to point the alias at")

		if err := flagset.Parse(args); err != nil {
			return err
		}
		if indexName == person.AliasName {
			aliasName = person.AliasName
		} else {
			aliasName = firm.AliasName
		}
		c.logger.Println("aliasName")
		c.logger.Println(aliasName)

		aliasedIndex, err := c.client.ResolveAlias(aliasName)

		if err != nil {
			return err
		}

		if aliasedIndex == *set {
			c.logger.Printf("alias '%s' is already set to '%s'", aliasName, *set)
			return nil
		}
		indexesToBeAliased = append (indexesToBeAliased, aliasedIndex)
	}

	for _, currentAlias := range indexesToBeAliased {
		err := c.client.UpdateAlias(aliasName, currentAlias, *set)
		if err != nil {
			return err
		}
	}
	return nil
}

//func (c *updateAliasCommand) Run(args []string) error {
//	l := logrus.New()
//	l.SetFormatter(&logrus.JSONFormatter{})
//	l.Println("Running update alias")
//	l.Println(c.index)
//
//	flagset := flag.NewFlagSet("update-alias", flag.ExitOnError)
//
//	indexName := strings.Split(c.index, "_")[0]
//
//	set := flagset.String("set", c.index, "index to point the alias at")
//
//	if err := flagset.Parse(args); err != nil {
//		return err
//	}
//	var aliasName string
//	if indexName == person.AliasName {
//		aliasName = person.AliasName
//	} else {
//		aliasName = firm.AliasName
//	}
//	c.logger.Println(aliasName)
//	aliasedIndex, err := c.client.ResolveAlias(aliasName)
//	if err != nil {
//		return err
//	}
//
//	if aliasedIndex == *set {
//		c.logger.Printf("alias '%s' is already set to '%s'", aliasName, *set)
//		return nil
//	}
//
//	return c.client.UpdateAlias(aliasName, aliasedIndex, *set)
//}
