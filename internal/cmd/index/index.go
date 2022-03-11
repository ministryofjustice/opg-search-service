package index

import (
	"context"
	"github.com/ministryofjustice/opg-search-service/internal/indices"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"strings"
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

type Results struct {
	Result *Result
	Error  error
}

type AllResults map[int]Results


func New(conn *pgx.Conn, es BulkClient, logger Logger, indexNames []string) *Indexer {
	return &Indexer{
		conn:      conn,
		es:        es,
		log:       logger,
		indexNames: indexNames,
	}
}

type Indexer struct {
	conn      *pgx.Conn
	es        BulkClient
	log       Logger
	indexNames [] string
}

func (r *Indexer) All(ctx context.Context, batchSize int) AllResults {
	var res *Result
	var allResults = make(map[int]Results)

	for i,indexName := range r.indexNames {
		aliasName := strings.Split(indexName, "_")[0]
		min, max, err := r.getIDRange(ctx, aliasName)

		if err != nil {
			return AllResults{
				0: {nil,err},
			}
		}

		res, err =  r.ByID(ctx, min, max, batchSize, aliasName)

		results := Results{
			Result: res,
			Error: err,
		}

		allResults[i] = results

	}
	return allResults

}

func (r *Indexer) ByID(ctx context.Context, start, end, batchSize int, aliasName string) (*Result, error) {
	var rerr error
	var result *Result
	var err error

	var index string
	for _, currentIndex := range r.indexNames {
		currentAliasName := strings.Split(currentIndex, "_")[0]
		if currentAliasName == aliasName {
			index = currentIndex
			break
		}
	}

	entity := make(chan indices.Entity, batchSize)
	go func() {
		err := r.queryByID(ctx, entity, start, end, batchSize, index, aliasName)
		if err != nil {
			rerr = err
		}
	}()
	result, err = r.index(ctx, entity, index)

	if rerr != nil {
		return result, rerr
	}

	return result, err
}

func (r *Indexer) FromDate(ctx context.Context, from time.Time, batchSize int) (*Result, error) {
	var rerr error
	entity := make(chan indices.Entity, batchSize)

	var personIndexName string
	for _,indexName := range r.indexNames {
		aliasName := strings.Split(indexName, "_")[0]
		if aliasName == person.AliasName {
			personIndexName = indexName
		}
	}

	go func() {
		err := r.queryFromDate(ctx, entity, from, personIndexName, person.AliasName)
		if err != nil {
			rerr = err
		}
	}()

	result, err := r.index(ctx, entity, personIndexName)
	if rerr != nil {
		return result, rerr
	}

	return result, err
}