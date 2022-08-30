package deputy

import (
	"encoding/json"

	"github.com/ministryofjustice/opg-search-service/internal/index"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	Deputies []Deputy `json:"persons"`
}

func (r *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if r.Deputies == nil || len(r.Deputies) == 0 {
		errs = append(errs, response.Error{
			Name:        "persons",
			Description: "field is empty",
		})
	}

	for _, p := range r.Deputies {
		errs = append(errs, p.Validate()...)
	}

	return errs
}

func (r *IndexRequest) Items() []index.Indexable {
	indexables := make([]index.Indexable, len(r.Deputies))
	for i, deputy := range r.Deputies {
		indexables[i] = deputy
	}

	return indexables
}

func ParseIndexRequest(body []byte) (index.Validatable, error) {
	var req IndexRequest
	err := json.Unmarshal(body, &req)

	return &req, err
}
