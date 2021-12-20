package cli

import (
	"context"
	"flag"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/internal/cmd/reindex"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
)

type reindexCommand struct {
	logger   *logrus.Logger
	esClient elasticsearch.ClientInterface
	exit     func(code int)

	shouldRun *bool
	db        *string
	from      *int
	to        *int
	batchSize *int
	fromDate  *string
}

func NewReindex(logger *logrus.Logger) *reindexCommand {
	esClient, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Fatal(err)
	}

	return &reindexCommand{
		logger:   logger,
		esClient: esClient,
		exit:     os.Exit,
	}
}

func (c *reindexCommand) DefineFlags() {
	c.shouldRun = flag.Bool("reindex", false, "reindex elasticsearch")
	c.db = flag.String("db", "", "database connection string")
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

	conn, err := pgx.Connect(ctx, *c.db)
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
