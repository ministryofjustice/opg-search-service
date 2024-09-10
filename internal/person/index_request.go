package person

import (
	"encoding/json"

	"github.com/ministryofjustice/opg-search-service/internal/index"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

type IndexRequest struct {
	Persons []Person `json:"persons"`
}

func (r *IndexRequest) Validate() []response.Error {
	var errs []response.Error

	if len(r.Persons) == 0 {
		errs = append(errs, response.Error{
			Name:        "persons",
			Description: "field is empty",
		})
	}

	for _, p := range r.Persons {
		errs = append(errs, p.Validate()...)
	}

	return errs
}

func (r *IndexRequest) Items() []index.Indexable {
	indexables := make([]index.Indexable, len(r.Persons))
	for i, person := range r.Persons {
		indexables[i] = person
	}

	return indexables
}

func ParseIndexRequest(body []byte) (index.Validatable, error) {
	var req IndexRequest
	err := json.Unmarshal(body, &req)

	return &req, err
}
