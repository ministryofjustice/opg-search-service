package indexing

import (
	"github.com/ministryofjustice/opg-search-service/internal/indices"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	Entities []indices.Entity
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if ir.Entities == nil || len(ir.Entities) == 0 {
		errs = append(errs, response.Error{
			Name:        "entity",
			Description: "field is empty",
		})
	}

	for _, p := range ir.Entities {
		errs = append(errs, p.Validate()...)
	}

	return errs
}
