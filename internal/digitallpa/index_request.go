package digitallpa

import (
	"encoding/json"

	"github.com/ministryofjustice/opg-search-service/internal/index"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	DigitalLpaApplications []DigitalLpa `json:"digitalLpaApplications"`
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if len(ir.DigitalLpaApplications) == 0 {
		errs = append(errs, response.Error{
			Name:        "digitalLpaApplications",
			Description: "field is empty",
		})
	}
	for _, p := range ir.DigitalLpaApplications {
		errs = append(errs, p.Validate()...)
	}

	return errs
}

func (r *IndexRequest) Items() []index.Indexable {
	indexables := make([]index.Indexable, len(r.DigitalLpaApplications))
	for i, digitalLpaApplication := range r.DigitalLpaApplications {
		indexables[i] = digitalLpaApplication
	}

	return indexables
}

func ParseIndexRequest(body []byte) (index.Validatable, error) {
	var req IndexRequest
	err := json.Unmarshal(body, &req)
	return &req, err
}
