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

func (r *Indexer) All(ctx context.Context, batchSize int) (*Result, error) {

	var tableName string
	var res *Result
	var err error

	for _,index := range r.indexNames {
		aliasName := strings.Split(index, "_")[0]
		switch aliasName {
		case indices.AliasNameFirm:
			tableName = indices.AliasNameFirm
		default:
			tableName = "persons"
		}

		min, max, err := r.getIDRange(ctx, tableName)

		r.log.Printf("in All index/index.go min", min)
		r.log.Printf("in All index/index.go max", max)
		r.log.Printf("in All index/index.go err", err)

		if err != nil {
			return nil, err
		}

		res, err =  r.ByID(ctx, min, max, batchSize, index)

		r.log.Printf("in All index/index.go res", res)
		r.log.Printf("in All index/index.go err", err)
	}
	return res, err

}

func (r *Indexer) ByID(ctx context.Context, start, end, batchSize int, aliasName string) (*Result, error) {
	var rerr error
	var result *Result
	var err error

	var index string
	for _, currentIndex := range r.indexNames {
		currentAliasName := strings.Split(index, "_")[0]
		//r.log.Printf("in index aliasName", aliasName)
		if currentAliasName == aliasName {
			index = currentIndex
		}
	}

	entity := make(chan indices.Entity, batchSize)
	go func() {

		//TODO some of these methods just need alias name to decide which query or scan to run when indexing - see where you
		//can add aliasname instead of the full index name
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