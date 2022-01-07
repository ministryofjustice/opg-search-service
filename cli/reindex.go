package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/internal/cmd/reindex"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

type Secrets interface {
	GetGlobalSecretString(key string) (string, error)
}

type reindexCommand struct {
	logger   *logrus.Logger
	esClient reindex.BulkClient
	secrets  Secrets
}

func NewReindex(logger *logrus.Logger, secrets Secrets) *reindexCommand {
	esClient, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Fatal(err)
	}

	return &reindexCommand{
		logger:   logger,
		esClient: esClient,
		secrets:  secrets,
	}
}

func (c *reindexCommand) Name() string {
	return "reindex"
}

func (c *reindexCommand) DefineFlags() {}

func (c *reindexCommand) ShouldRun() bool {
	return false
}

func (c *reindexCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("reindex", flag.ExitOnError)

	from := flagset.Int("from", 0, "id to index from")
	to := flagset.Int("to", 100, "id to index to")
	batchSize := flagset.Int("batch-size", 1000, "batch size to read from db")
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

	reindexer := reindex.New(conn, c.esClient, c.logger)

	fromTime, err := time.Parse(time.RFC3339, *fromDate)

	if *fromDate != "" && err != nil {
		return fmt.Errorf("-from-date: %w", err)
	}

	var result *reindex.Result

	if !fromTime.IsZero() {
		c.logger.Printf("indexing by date from=%v", fromDate)
		result, err = reindexer.ByDate(ctx, fromTime)
	} else {
		c.logger.Printf("indexing by id from=%d to=%d batchSize=%d", *from, *to, *batchSize)
		result, err = reindexer.ByID(ctx, *from, *to, *batchSize)
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

func (c *reindexCommand) dbConnectionString() (string, error) {
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
