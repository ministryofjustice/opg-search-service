package index

import (
	"context"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/person"
)

type BulkClient interface {
	DoBulk(*elasticsearch.BulkOp) (elasticsearch.BulkResult, error)
}

type Logger interface {
	Printf(string, ...interface{})
}

func New(conn *pgx.Conn, es BulkClient, logger Logger, indexName string) *Indexer {
	return &Indexer{
		conn:      conn,
		es:        es,
		log:       logger,
		indexName: indexName,
	}
}

type Indexer struct {
	conn      *pgx.Conn
	es        BulkClient
	log       Logger
	indexName string
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

func (r *Indexer) FromDatePerson(ctx context.Context, from time.Time, batchSize int) (*Result, error) {
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

func (r *Indexer) FromDateFirm(ctx context.Context, from time.Time, batchSize int) (*Result, error) {
	var rerr error
	firms := make(chan firm.Firm, batchSize)

	go func() {
		err := r.queryFromDate(ctx, firms, from)
		if err != nil {
			rerr = err
		}
	}()

	result, err := r.index(ctx, firms)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}
