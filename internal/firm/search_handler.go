package firm

import (
	"encoding/json"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type SearchClient interface {
	Search(indexName string, requestBody map[string]interface{}) (*elasticsearch.SearchResult, error)
}

type SearchHandler struct {
	logger *logrus.Logger
	client SearchClient
}

func NewSearchHandler(logger *logrus.Logger, client SearchClient) *SearchHandler {
	return &SearchHandler{
		logger: logger,
		client: client,
	}
}

func (s *SearchHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.Println("in firm search handler")

	req, err := person.CreateSearchRequestFromRequest(r)
	s.logger.Println("after create search request")

	if err != nil {
		s.logger.Println(err)
		response.WriteJSONError(rw, "request", err.Error(), http.StatusBadRequest)
		return
	}

	filters := make([]interface{}, 0)
	for _, f := range req.PersonTypes {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{
				"personType": f,
			},
		})
	}
	s.logger.Println("before es req body")

	// construct ES request body
	esReqBody := map[string]interface{}{
		"from": req.From,
		//"sort": map[string]interface{}{
		//	"firmName": map[string]string{
		//		"order": "asc",
		//	},
		//},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": req.Term,
						"fields": []string{
							"firmName",
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]string{
					"field": "personType.keyword",
				},
			},
		},
		"post_filter": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": filters,
			},
		},
	}

	if req.Size > 0 {
		esReqBody["size"] = req.Size
	}

	result, err := s.client.Search(AliasName, esReqBody)
	s.logger.Println("after result")
	s.logger.Println(result)

	if err != nil {
		s.logger.Println(err.Error())
		response.WriteJSONError(rw, "request", "Firm search caused an unexpected error", http.StatusInternalServerError)
		return
	}

	resp := response.SearchResponse{
		Aggregations: result.Aggregations,
		Results:      result.Hits,
		Total: response.Total{
			Count: result.Total,
			Exact: result.TotalExact,
		},
	}

	jsonResp, _ := json.Marshal(resp)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(jsonResp)
	s.logger.Println("json response")
	s.logger.Println(jsonResp)


	s.logger.Printf("Request took: %d", time.Since(start))
}
