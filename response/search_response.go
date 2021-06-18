package response

import "opg-search-service/elasticsearch"

type SearchResponse struct {
	Results      []elasticsearch.Indexable `json:"results"`
	Aggregations map[string]map[string]int `json:"aggregations"`
	Total        Total                     `json:"total"`
}

type Total struct {
	Count int  `json:"count"`
	Exact bool `json:"exact"`
}
