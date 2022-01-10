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

func (r *Indexer) All(ctx context.Context, batchSize int) (*Result, error) {
	min, max, err := r.getIDRange(ctx)
	if err != nil {
		return nil, err
	}

	return r.ByID(ctx, min, max, batchSize)
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

func (r *Indexer) FromDate(ctx context.Context, from time.Time, batchSize int) (*Result, error) {
	var rerr error
	persons := make(chan person.Person, batchSize)

	go func() {
		err := r.queryFromDate(ctx, persons, from)
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
