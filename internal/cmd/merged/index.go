package merged

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

type createIndexCommand struct {
	commands []*indexCommand
}

func NewIndex(logger *logrus.Logger, esClient index.BulkClient, secrets Secrets, indexes map[string][]byte) *createIndexCommand {
	commandArray := &createIndexCommand{}

	for indexName, _ := range indexes {
		indexCommand := &indexCommand{
			logger:    logger,
			esClient:  esClient,
			secrets:   secrets,
			indexName: indexName,
		}
		commandArray.commands = append(commandArray.commands, indexCommand)
	}
	return commandArray
}

func (c *createIndexCommand) Name() string {
	return "index"
}

func (c *createIndexCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("index", flag.ExitOnError)

	all := flagset.Bool("all", false, "index all records")
	person := flagset.Bool("person", false, "index person records")
	firm := flagset.Bool("firm", false, "index firm records")
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

	for _, s := range c.commands {
		indexer := index.New(conn, s.esClient, s.logger, s.indexName)

		fromTime, err := time.Parse(time.RFC3339, *fromDate)

		if *fromDate != "" && err != nil {
			return fmt.Errorf("-from-date: %w", err)
		}

		var result *index.Result
		indexName := strings.Split(s.indexName, "_")[0]
		if !fromTime.IsZero() && indexName == "person" {
			s.logger.Printf("indexing by date from=%v batchSize=%d", fromTime, *batchSize)
			result, err = indexer.FromDate(ctx, fromTime, *batchSize)
		} else if *all {
			s.logger.Printf("indexing all records batchSize=%d", *batchSize)
			result, err = indexer.All(ctx, *batchSize, indexName)
		} else if *person && indexName == "person" {
			s.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
			result, err = indexer.ByID(ctx, *from, *to, *batchSize)
		} else if *firm && indexName == "firm" {
			s.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
			result, err = indexer.ByID(ctx, *from, *to, *batchSize)
		}

		if err != nil {
			return err
		}

		s.logger.Printf("indexing done successful=%d failed=%d", result.Successful, result.Failed)
		for _, e := range result.Errors {
			s.logger.Println(e)
		}
	}
	return nil
}

func (c *createIndexCommand) dbConnectionString() (string, error) {
	pass := os.Getenv("SEARCH_SERVICE_DB_PASS")
	universalSecret := c.commands[0].secrets
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
