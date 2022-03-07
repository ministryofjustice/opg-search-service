package indices

import (
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	//tried to set this to an array of entities but as it has no idea what to unmarshal the json as - it fails
	Firms []Firm `json:"firms"`
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if ir.Firms == nil || len(ir.Firms) == 0 {
		errs = append(errs, response.Error{
			Name:        "entity",
			Description: "field is empty",
		})
	}

	fmt.Println("before for loop")

	for _, p := range ir.Firms {
		errs = append(errs, p.Validate()...)
	}

	return errs
}
