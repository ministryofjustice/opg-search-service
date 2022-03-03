package index

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/Merged"
	"github.com/sirupsen/logrus"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
)

func (r *Indexer) indexPerson(ctx context.Context, persons <-chan Merged.Person) (*Result, error) {
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

func (r *Indexer) indexFirm(ctx context.Context, firms <-chan Merged.Firm) (*Result, error) {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.Println("in index firm")
	l.Println("index name", r.indexName)
	op := elasticsearch.NewBulkOp(r.indexName)
	result := &Result{}

	for f := range firms {
		f.Persontype = "Firm"
		err := op.Index(f.Id(), f)

		if err == elasticsearch.ErrOpTooLarge {
			res, bulkErr := r.es.DoBulk(op)
			if bulkErr == nil {
				r.log.Printf("batch indexed successful=%d failed=%d error=%s", res.Successful, res.Failed, res.Error)
			} else {
				r.log.Printf("indexing error: %s", bulkErr.Error())
			}

			result.Add(res, bulkErr)
			op.Reset()
			err = op.Index(f.Id(), f)
		}

		if err != nil {
			return nil, fmt.Errorf("could not construct index request for id=%d; %w", f.Id(), err)
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
