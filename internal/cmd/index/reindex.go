package index

import (
	"context"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"time"

	"github.com/jackc/pgx/v4"
)

type BulkClient interface {
	DoBulk(*elasticsearch.BulkOp) (elasticsearch.BulkResult, error)
}

type Logger interface {
	Printf(string, ...interface{})
}

func New(conn *pgx.Conn, es BulkClient, logger Logger) *Indexer {
	return &Indexer{
		conn: conn,
		es:   es,
		log:  logger,
	}
}

type Indexer struct {
	conn *pgx.Conn
	es   BulkClient
	log  Logger
}

func (r *Indexer) ByID(ctx context.Context, start, end, batchSize int) (*Result, error) {
	var rerr error
	persons := make(chan person.Person, batchSize)

	go func() {
		err := r.queryByID(ctx, persons, start, end, batchSize)
		if err != nil {
			rerr = err
		}
	}()

	result, err := r.index(ctx, persons)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}

func (r *Indexer) ByDate(ctx context.Context, from time.Time) (*Result, error) {
	var rerr error
	persons := make(chan person.Person, 100)

	go func() {
		err := r.queryByDate(ctx, persons, from)
		if err != nil {
			rerr = err
		}
	}()

	result, err := r.index(ctx, persons)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}
