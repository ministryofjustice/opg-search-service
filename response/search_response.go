package response

import "opg-search-service/elasticsearch"

type SearchResponse struct {
	Results []elasticsearch.Indexable `json:"results"`
}
