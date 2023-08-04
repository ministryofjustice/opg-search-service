package index

import (
	"context"
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
)

type DB interface {
	QueryIDRange(ctx context.Context) (min, max int, err error)
	QueryByID(ctx context.Context, results chan<- Indexable, from, to int) error
	QueryFromDate(ctx context.Context, results chan<- Indexable, fromDate time.Time) error
}

type BulkClient interface {
	DoBulk(ctx context.Context, op *elasticsearch.BulkOp) (elasticsearch.BulkResult, error)
}

type Logger interface {
	Printf(string, ...interface{})
}

func New(es BulkClient, logger Logger, db DB, indexName string) *Indexer {
	return &Indexer{
		es:        es,
		log:       logger,
		db:        db,
		indexName: indexName,
	}
}

type Indexer struct {
	es        BulkClient
	log       Logger
	db        DB
	indexName string
}

func (r *Indexer) All(ctx context.Context, batchSize int) (*Result, error) {
	min, max, err := r.db.QueryIDRange(ctx)
	if err != nil {
		return nil, err
	}

	return r.ByID(ctx, min, max, batchSize)
}

func (r *Indexer) ByID(ctx context.Context, start, end, batchSize int) (*Result, error) {
	var rerr error
	items := make(chan Indexable, batchSize)

	go func() {
		defer func() { close(items) }()

		batch := &batchIter{start: start, end: end, size: batchSize}

		for batch.Next() {
			r.log.Printf("reading range from db (%d, %d)", batch.From(), batch.To())

			if err := r.db.QueryByID(ctx, items, batch.From(), batch.To()); err != nil {
				rerr = err
				break
			}
		}
	}()

	result, err := r.index(ctx, items)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}

func (r *Indexer) FromDate(ctx context.Context, from time.Time, batchSize int) (*Result, error) {
	var rerr error
	items := make(chan Indexable, batchSize)

	go func() {
		defer func() { close(items) }()

		rerr = r.db.QueryFromDate(ctx, items, from)
	}()

	result, err := r.index(ctx, items)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}

func (r *Indexer) index(ctx context.Context, entity <-chan Indexable) (*Result, error) {
	op := elasticsearch.NewBulkOp(r.indexName)
	result := &Result{}

	for e := range entity {
		err := op.Index(e.Id(), e)

		if err == elasticsearch.ErrOpTooLarge {
			res, bulkErr := r.es.DoBulk(ctx, op)
			if bulkErr == nil {
				r.log.Printf("batch indexed successful=%d failed=%d error=%s", res.Successful, res.Failed, res.Error)
			} else {
				r.log.Printf("indexing error: %s", bulkErr.Error())
			}

			result.Add(res, bulkErr)
			op.Reset()
			err = op.Index(e.Id(), e)
		}

		if err != nil {
			return nil, fmt.Errorf("could not construct index request for id=%d; %w", e.Id(), err)
		}
	}

	if !op.Empty() {
		result.Add(r.es.DoBulk(ctx, op))
	}

	return result, nil
}

type Result struct {
	Successful int
	Failed     int
	Errors     []string
}

func (r *Result) Add(result elasticsearch.BulkResult, err error) {
	r.Successful += result.Successful
	r.Failed += result.Failed

	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
}
