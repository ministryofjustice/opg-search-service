package firm

import (
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	Firms []Firm `json:"firms"`
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if ir.Firms == nil || len(ir.Firms) == 0 {
		errs = append(errs, response.Error{
			Name:        "firms",
			Description: "field is empty",
		})
	}

	for _, f := range ir.Firms {
		errs = append(errs, f.Validate()...)
	}

	return errs
}
