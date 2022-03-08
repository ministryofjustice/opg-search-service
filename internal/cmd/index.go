package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/indices"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/cmd/index"
	"github.com/sirupsen/logrus"
)

type Secrets interface {
	GetGlobalSecretString(key string) (string, error)
}

type indexCommand struct {
	logger    		*logrus.Logger
	esClient  		index.BulkClient
	secrets   		Secrets
	currentIndices	[] string
}

func NewIndex(logger *logrus.Logger, esClient index.BulkClient, secrets Secrets, indexes map[string][]byte) *indexCommand {
	var currentIndices [] string
	for indexName, _ := range indexes {
		currentIndices = append(currentIndices, indexName)
	}
	return &indexCommand{
			logger:    logger,
			esClient:  esClient,
			secrets:   secrets,
			currentIndices: currentIndices,
	}
}

func (c *indexCommand) Name() string {
	return "index"
}

func (c *indexCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("index", flag.ExitOnError)

	all := flagset.Bool("all", false, "index all records")
	personOnly := flagset.Bool("person", false, "index person records")
	firmOnly := flagset.Bool("firm", false, "index firm records")
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

	var fromTime time.Time
	indexer := index.New(conn, c.esClient, c.logger, c.currentIndices)

	c.logger.Printf("error here")
	c.logger.Printf(*fromDate)

	fromTime, err = time.Parse(time.RFC3339, *fromDate)

	c.logger.Printf("2")
	c.logger.Printf(*fromDate)

	if *fromDate != "" && err != nil {
		return fmt.Errorf("-from-date: %w", err)
	}

	var result *index.Result

	if !fromTime.IsZero() {
		c.logger.Printf("indexing by date from=%v batchSize=%d", fromTime, *batchSize)
		result, err = indexer.FromDate(ctx, fromTime, *batchSize)
	} else if *all {
		c.logger.Printf("indexing all records batchSize=%d", *batchSize)
		result, err = indexer.All(ctx, *batchSize)
	} else if *personOnly {
		c.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
		result, err = indexer.ByID(ctx, *from, *to, *batchSize, person.AliasName)
	} else if *firmOnly {
		c.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
		result, err = indexer.ByID(ctx, *from, *to, *batchSize, indices.AliasNameFirm)
	}

	if err != nil {
		return err
	}

	c.logger.Printf("indexing done successful=%d failed=%d", result.Successful, result.Failed)
	for _, e := range result.Errors {
		c.logger.Println(e)
	}

	return nil
}

func (c *indexCommand) dbConnectionString() (string, error) {
	pass := os.Getenv("SEARCH_SERVICE_DB_PASS")
	universalSecret := c.secrets
	if passSecret := os.Getenv("SEARCH_SERVICE_DB_PASS_SECRET"); passSecret != "" {
		var err error
		pass, err = universalSecret.GetGlobalSecretString(passSecret)
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
