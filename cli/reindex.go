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
	GetSecretString(key string) (string, error)
}

type reindexCommand struct {
	logger   *logrus.Logger
	esClient reindex.BulkClient
	secrets  Secrets
	exit     func(code int)

	shouldRun *bool
	db        *string
	from      *int
	to        *int
	batchSize *int
	fromDate  *string
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
		exit:     os.Exit,
	}
}

func (c *reindexCommand) DefineFlags() {
	c.shouldRun = flag.Bool("reindex", false, "reindex elasticsearch")
	c.from = flag.Int("from", 0, "id to index from")
	c.to = flag.Int("to", 100, "id to index to")
	c.batchSize = flag.Int("batch-size", 1000, "batch size to read from db")
	c.fromDate = flag.String("from-date", "", "index all records updated from this date")
}

func (c *reindexCommand) ShouldRun() bool {
	return *c.shouldRun
}

func (c *reindexCommand) Run() {
	ctx := context.Background()

	connString, err := c.dbConnectionString()
	if err != nil {
		c.logger.Errorln(err)
		c.exit(1)
		return
	}

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		c.logger.Errorln(err)
		c.exit(1)
		return
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		c.logger.Errorln(err)
		c.exit(1)
		return
	}

	reindexer := reindex.New(conn, c.esClient)

	fromDate, err := time.Parse(time.RFC3339, *c.fromDate)

	if *c.fromDate != "" && err != nil {
		c.logger.Errorln("-from-date:", err)
		c.exit(1)
		return
	}

	var result *reindex.Result

	if !fromDate.IsZero() {
		result, err = reindexer.ByDate(ctx, fromDate)
	} else {
		result, err = reindexer.ByID(ctx, *c.from, *c.to, *c.batchSize)
	}

	if err != nil {
		c.logger.Errorln(err)
		c.exit(1)
		return
	}

	c.logger.Printf("indexing done successful=%d failed=%d", result.Successful, result.Failed)
	for _, e := range result.Errors {
		c.logger.Println(e)
	}

	c.exit(0)
}

func (c *reindexCommand) dbConnectionString() (string, error) {
	pass := os.Getenv("SEARCH_SERVICE_DB_PASS")
	if passSecret := os.Getenv("SEARCH_SERVICE_DB_PASS_SECRET"); passSecret != "" {
		var err error
		pass, err = c.secrets.GetSecretString(passSecret)
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

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, host, pass, port, database), nil
}
