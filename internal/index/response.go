package index

import "github.com/ministryofjustice/opg-search-service/internal/elasticsearch"

type indexResponse struct {
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors,omitempty"`
}

func (r *indexResponse) Add(result elasticsearch.BulkResult, err error) {
	r.Successful += result.Successful
	r.Failed += result.Failed

	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
}
