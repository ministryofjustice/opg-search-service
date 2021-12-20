package reindex

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

func New(conn *pgx.Conn, es BulkClient) *Reindexer {
	return &Reindexer{
		conn: conn,
		es:   es,
	}
}

type Reindexer struct {
	conn *pgx.Conn
	es   BulkClient
}

func (r *Reindexer) ByID(ctx context.Context, start, end, batchSize int) (*Result, error) {
	var rerr error
	persons := make(chan person.Person, batchSize)

	go func() {
		err := r.queryByID(ctx, persons, start, end, batchSize)
		if err != nil {
			rerr = err
		}
	}()

	result, err := r.reindex(ctx, persons)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}

func (r *Reindexer) ByDate(ctx context.Context, from time.Time) (*Result, error) {
	var rerr error
	persons := make(chan person.Person, 100)

	go func() {
		err := r.queryByDate(ctx, persons, from)
		if err != nil {
			rerr = err
		}
	}()

	result, err := r.reindex(ctx, persons)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}
