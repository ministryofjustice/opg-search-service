package person

import (
	"encoding/json"
	"errors"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/response"
	"time"

	"github.com/sirupsen/logrus"
)

type SearchHandler struct {
	logger *logrus.Logger
	es     elasticsearch.ClientInterface
}

func NewSearchHandler(logger *logrus.Logger) (*SearchHandler, error) {
	client, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Println(err)
		return nil, errors.New("unable to create a new Elasticsearch client")
	}

	return &SearchHandler{
		logger,
		client,
	}, nil
}

func (s SearchHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	req, err := CreateSearchRequestFromRequest(r)
	if err != nil {
		s.logger.Println(err)
		response.WriteJSONError(rw, "request", err.Error(), http.StatusBadRequest)
		return
	}

	var dataType Person

	filters := make([]interface{}, 0)
	for _, f := range req.PersonTypes {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{
				"personType": f,
			},
		})
	}

	// construct ES request body
	esReqBody := map[string]interface{}{
		"from": req.From,
		"sort": map[string]interface{}{
			"surname": map[string]string{
				"order": "asc",
			},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": req.Term,
						"fields": []string{
							"searchable",
							"caseRecNumber",
						},
						"default_operator": "AND",
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]string{
					"field": "personType",
				},
			},
		},
		"post_filter":  map[string]interface{}{
			"bool": map[string]interface{}{
				"should": filters,
			},
		},
	}

	if req.Size > 0 {
		esReqBody["size"] = req.Size
	}

	result, err := s.es.Search(esReqBody, dataType)
	if err != nil {
		s.logger.Println(err.Error())
		response.WriteJSONError(rw, "request", "Person search caused an unexpected error", http.StatusInternalServerError)
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

	s.logger.Printf("Request took: %d", time.Since(start))
}
