package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/cmd/index"
	"github.com/sirupsen/logrus"
)

type Secrets interface {
	GetGlobalSecretString(key string) (string, error)
}

type indexCommand struct {
	logger    *logrus.Logger
	esClient  index.BulkClient
	secrets   Secrets
	indexName string
}

type indexCommandForPersonAndFirm struct {
	logger    		*logrus.Logger
	esClient  		index.BulkClient
	secrets   		Secrets
	personIndexName string
	firmIndexName	string
}

func NewIndex(logger *logrus.Logger, esClient index.BulkClient, secrets Secrets, indexName string) *indexCommand {
	return &indexCommand{
		logger:    logger,
		esClient:  esClient,
		secrets:   secrets,
		indexName: indexName,
	}
}

func NewIndexForPersonAndFirm(logger *logrus.Logger, esClient index.BulkClient, secrets Secrets, personIndexName string, firmIndexName string) *indexCommandForPersonAndFirm {
	return &indexCommandForPersonAndFirm{
		logger:    logger,
		esClient:  esClient,
		secrets:   secrets,
		personIndexName: personIndexName,
		firmIndexName: firmIndexName,
	}
}

//func (c *indexCommand) Name() string {
//	return "index"
//}

func (c *indexCommandForPersonAndFirm) Name() string {
	return "index"
}

func (c *indexCommandForPersonAndFirm) Run(args []string) error {
	flagset := flag.NewFlagSet("index", flag.ExitOnError)

	all := flagset.Bool("all", false, "index all records")
	from := flagset.Int("from", 0, "id to index from")
	to := flagset.Int("to", 100, "id to index to")
	batchSize := flagset.Int("batch-size", 10000, "batch size to read from db")
	fromDate := flagset.String("from-date", "", "index all records updated from this date")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	ctx := context.Background()

	connString, err := c.dbConnectionString()
	if err != nil {
		return err
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		return err
	}

	var indexes = []string{c.personIndexName, c.firmIndexName}

	for _, currentIndex := range indexes {
		indexer := index.New(conn, c.esClient, c.logger, currentIndex)
		fromTime, err := time.Parse(time.RFC3339, *fromDate)

		if *fromDate != "" && err != nil {
			return fmt.Errorf("-from-date: %w", err)
		}

		var result *index.Result
		indexName := strings.Split(currentIndex, "_")[0]
		if !fromTime.IsZero() {
			c.logger.Printf("indexing by date from=%v batchSize=%d", fromTime, *batchSize)
			result, err = indexer.FromDatePerson(ctx, fromTime, *batchSize)
			//
			//if c.indexName == "person" {
			//	result, err = indexer.FromDatePerson(ctx, fromTime, *batchSize)
			//} else {
			//	result, err = indexer.FromDateFirm(ctx, fromTime, *batchSize)
			//}
		} else if *all {
			c.logger.Printf("indexing all records batchSize=%d", *batchSize)
			c.logger.Println("index name", currentIndex)
			c.logger.Println("context", ctx)

			result, err = indexer.All(ctx, *batchSize, currentIndex)
		} else {
			c.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
			if indexName == "person" {
				result, err = indexer.ByIDPerson(ctx, *from, *to, *batchSize)
			} else {
				result, err = indexer.ByIDFirm(ctx, *from, *to, *batchSize)
			}
		}

		if err != nil {
			return err
		}

		c.logger.Printf("indexing done successful=%d failed=%d", result.Successful, result.Failed)
		for _, e := range result.Errors {
			c.logger.Println(e)
		}
	}
	return nil
}

//func (c *indexCommand) Run(args []string) error {
//	flagset := flag.NewFlagSet("index", flag.ExitOnError)
//
//	all := flagset.Bool("all", false, "index all records")
//	from := flagset.Int("from", 0, "id to index from")
//	to := flagset.Int("to", 100, "id to index to")
//	batchSize := flagset.Int("batch-size", 10000, "batch size to read from db")
//	fromDate := flagset.String("from-date", "", "index all records updated from this date")
//
//	if err := flagset.Parse(args); err != nil {
//		return err
//	}
//
//	ctx := context.Background()
//
//	connString, err := c.dbConnectionString()
//	if err != nil {
//		return err
//	}
//
//	conn, err := pgx.Connect(ctx, connString)
//	if err != nil {
//		return err
//	}
//	defer conn.Close(ctx)
//
//	if err := conn.Ping(ctx); err != nil {
//		return err
//	}
//
//	indexer := index.New(conn, c.esClient, c.logger, c.indexName)
//
//	fromTime, err := time.Parse(time.RFC3339, *fromDate)
//
//	if *fromDate != "" && err != nil {
//		return fmt.Errorf("-from-date: %w", err)
//	}
//
//	var result *index.Result
//	index := strings.Split(c.indexName, "_")[0]
//
//	if !fromTime.IsZero() {
//		c.logger.Printf("indexing by date from=%v batchSize=%d", fromTime, *batchSize)
//		result, err = indexer.FromDatePerson(ctx, fromTime, *batchSize)
//		//
//		//if c.indexName == "person" {
//		//	result, err = indexer.FromDatePerson(ctx, fromTime, *batchSize)
//		//} else {
//		//	result, err = indexer.FromDateFirm(ctx, fromTime, *batchSize)
//		//}
//	} else if *all {
//		c.logger.Printf("indexing all records batchSize=%d", *batchSize)
//		c.logger.Println("index name", c.indexName)
//		c.logger.Println("context", ctx)
//
//		result, err = indexer.All(ctx, *batchSize, c.indexName)
//	} else {
//		c.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
//		if index == "person" {
//			result, err = indexer.ByIDPerson(ctx, *from, *to, *batchSize)
//		} else {
//			result, err = indexer.ByIDFirm(ctx, *from, *to, *batchSize)
//		}
//	}
//
//	if err != nil {
//		return err
//	}
//
//	c.logger.Printf("indexing done successful=%d failed=%d", result.Successful, result.Failed)
//	for _, e := range result.Errors {
//		c.logger.Println(e)
//	}
//
//	return nil
//}

func (c *indexCommandForPersonAndFirm) dbConnectionString() (string, error) {
	pass := os.Getenv("SEARCH_SERVICE_DB_PASS")
	if passSecret := os.Getenv("SEARCH_SERVICE_DB_PASS_SECRET"); passSecret != "" {
		var err error
		pass, err = c.secrets.GetGlobalSecretString(passSecret)
		if err != nil {
			return "", err
		}
	}
	if pass == "" {
		return "", errors.New("SEARCH_SERVICE_DB_PASS or SEARCH_SERVICE_DB_PASS_SECRET must be specified")
	}

	user, host, port, database := os.Getenv("SEARCH_SERVICE_DB_USER"), os.Getenv("SEARCH_SERVICE_DB_HOST"), os.Getenv("SEARCH_SERVICE_DB_PORT"), os.Getenv("SEARCH_SERVICE_DB_DATABASE")
	if user == "" {
		return "", errors.New("SEARCH_SERVICE_DB_USER must be specified")
	}
	if host == "" {
		return "", errors.New("SEARCH_SERVICE_DB_HOST must be specified")
	}
	if port == "" {
		return "", errors.New("SEARCH_SERVICE_DB_PORT must be specified")
	}
	if database == "" {
		return "", errors.New("SEARCH_SERVICE_DB_DATABASE must be specified")
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, database), nil
}

//func (c *indexCommand) dbConnectionString() (string, error) {
//	pass := os.Getenv("SEARCH_SERVICE_DB_PASS")
//	if passSecret := os.Getenv("SEARCH_SERVICE_DB_PASS_SECRET"); passSecret != "" {
//		var err error
//		pass, err = c.secrets.GetGlobalSecretString(passSecret)
//		if err != nil {
//			return "", err
//		}
//	}
//	if pass == "" {
//		return "", errors.New("SEARCH_SERVICE_DB_PASS or SEARCH_SERVICE_DB_PASS_SECRET must be specified")
//	}
//
//	user, host, port, database := os.Getenv("SEARCH_SERVICE_DB_USER"), os.Getenv("SEARCH_SERVICE_DB_HOST"), os.Getenv("SEARCH_SERVICE_DB_PORT"), os.Getenv("SEARCH_SERVICE_DB_DATABASE")
//	if user == "" {
//		return "", errors.New("SEARCH_SERVICE_DB_USER must be specified")
//	}
//	if host == "" {
//		return "", errors.New("SEARCH_SERVICE_DB_HOST must be specified")
//	}
//	if port == "" {
//		return "", errors.New("SEARCH_SERVICE_DB_PORT must be specified")
//	}
//	if database == "" {
//		return "", errors.New("SEARCH_SERVICE_DB_DATABASE must be specified")
//	}
//
//	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, pass, host, port, database), nil
//}
