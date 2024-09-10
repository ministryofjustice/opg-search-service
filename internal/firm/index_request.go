package firm

import (
	"encoding/json"

	"github.com/ministryofjustice/opg-search-service/internal/index"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	Firms []Firm `json:"firms"`
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if len(ir.Firms) == 0 {
		errs = append(errs, response.Error{
			Name:        "entity",
			Description: "field is empty",
		})
	}
	for _, p := range ir.Firms {
		errs = append(errs, p.Validate()...)
	}

	return errs
}

func (r *IndexRequest) Items() []index.Indexable {
	indexables := make([]index.Indexable, len(r.Firms))
	for i, firm := range r.Firms {
		indexables[i] = firm
	}

	return indexables
}

func ParseIndexRequest(body []byte) (index.Validatable, error) {
	var req IndexRequest
	err := json.Unmarshal(body, &req)

	return &req, err
}
