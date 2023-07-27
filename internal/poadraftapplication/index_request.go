package poadraftapplication

import (
	"encoding/json"

	"github.com/ministryofjustice/opg-search-service/internal/index"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	DraftApplications []DraftApplication `json:"draftApplications"`
}

func (ir *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if ir.DraftApplications == nil || len(ir.DraftApplications) == 0 {
		errs = append(errs, response.Error{
			Name: "entity",
			Description: "field is empty",
		})
	}
	for _, p := range ir.DraftApplications {
		errs = append(errs, p.Validate()...)
	}

	return errs
}

func (r *IndexRequest) Items() []index.Indexable {
	indexables := make([]index.Indexable, len(r.DraftApplications))
	for i, draftApplication := range r.DraftApplications {
		indexables[i] = draftApplication
	}

	return indexables
}

func ParseIndexRequest(body []byte) (index.Validatable, error) {
	var req IndexRequest
	err := json.Unmarshal(body, &req)
	return &req, err
}
