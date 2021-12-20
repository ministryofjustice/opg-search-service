package reindex

import (
	"context"
	"fmt"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
)

func (r *Reindexer) reindex(ctx context.Context, persons <-chan person.Person) (*Result, error) {
	op := elasticsearch.NewBulkOp("person")
	result := &Result{}

	for p := range persons {
		err := op.Index(p.Id(), p)

		if err == elasticsearch.ErrOpTooLarge {
			result.Add(r.es.DoBulk(op))
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
