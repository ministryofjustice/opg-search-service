package person

import (
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	Persons []Person `json:"persons"`
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if ir.Persons == nil || len(ir.Persons) == 0 {
		errs = append(errs, response.Error{
			Name:        "persons",
			Description: "field is empty",
		})
	}

	for _, p := range ir.Persons {
		errs = append(errs, p.Validate()...)
	}

	return errs
}
