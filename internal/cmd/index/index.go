package index

import (
	"context"
	"github.com/ministryofjustice/opg-search-service/internal/indices"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
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

func (r *Indexer) All(ctx context.Context, batchSize int, indexName string) (*Result, error) {
	var tableName string

	switch indexName {
	case indices.AliasNameFirm:
		tableName = indices.AliasNameFirm
	default:
		tableName = "persons"
	}

	r.log.Printf(indexName)

	min, max, err := r.getIDRange(ctx, tableName)

	if err != nil {
		return nil, err
	}

	return r.ByID(ctx, min, max, batchSize, indexName)
}

func (r *Indexer) ByID(ctx context.Context, start, end, batchSize int, indexName string) (*Result, error) {
	var rerr error
	var result *Result
	var err error

	r.log.Printf("indexName")
	r.log.Printf(indexName)
	switch indexName {
	case indices.AliasNameFirm:
		firms := make(chan indices.Firm, batchSize)
		go func() {
			err := r.queryByIDFirm(ctx, firms, start, end, batchSize)
			if err != nil {
				rerr = err
			}
		}()
		result, err = r.indexFirm(ctx, firms)
	default:
		persons := make(chan person.Person, batchSize)
		go func() {
			err := r.queryByIDPerson(ctx, persons, start, end, batchSize)
			if err != nil {
				rerr = err
			}
		}()
		result, err = r.indexPerson(ctx, persons)
	}

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

	result, err := r.indexPerson(ctx, persons)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}