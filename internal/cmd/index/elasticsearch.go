package index

import (
	"context"
	"fmt"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
)

func (r *Indexer) index(ctx context.Context, persons <-chan person.Person) (*Result, error) {
	op := elasticsearch.NewBulkOp(r.indexName)
	result := &Result{}

	for p := range persons {
		err := op.Index(p.Id(), p)

		if err == elasticsearch.ErrOpTooLarge {
			res, bulkErr := r.es.DoBulk(op)
			if bulkErr == nil {
				r.log.Printf("batch indexed successful=%d failed=%d error=%s", res.Successful, res.Failed, res.Error)
			} else {
				r.log.Printf("indexing error: %s", bulkErr.Error())
			}

			result.Add(res, bulkErr)
			op.Reset()
			err = op.Index(p.Id(), p)
		}

		if err != nil {
			return nil, fmt.Errorf("could not construct index request for id=%d; %w", p.Id(), err)
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
