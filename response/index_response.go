package response

import "opg-search-service/elasticsearch"

type IndexResponse struct {
	Results []elasticsearch.IndexResult
}
