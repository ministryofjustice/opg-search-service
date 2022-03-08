package index

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/indices"
	"strings"
)

func (r *Indexer) index(ctx context.Context, entity <-chan indices.Entity, indexName string) (*Result, error) {

		aliasOfIndex := strings.Split(indexName, "_")[0]

		//r.log.Printf("in index indexname", indexName)
		//r.log.Printf("in index aliasOfIndex", aliasOfIndex)

	var indexToIndex string
	for _, index := range r.indexNames {
		aliasName := strings.Split(index, "_")[0]
		//r.log.Printf("in index aliasName", aliasName)
		if aliasName == aliasOfIndex {
			indexToIndex = index
		}
	}
	//r.log.Printf("in index indextoindex", indexToIndex)

	op := elasticsearch.NewBulkOp(indexToIndex)
	result := &Result{}

	for e := range entity {
		err := op.Index(e.Id(), e)

		if err == elasticsearch.ErrOpTooLarge {
			res, bulkErr := r.es.DoBulk(op)
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
		result.Add(r.es.DoBulk(op))
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
