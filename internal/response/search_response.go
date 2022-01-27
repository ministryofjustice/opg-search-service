package response

import (
	"encoding/json"
)

type SearchResponse struct {
	Results      []json.RawMessage         `json:"results"`
	Aggregations map[string]map[string]int `json:"aggregations"`
	Total        Total                     `json:"total"`
}

type Total struct {
	Count int  `json:"count"`
	Exact bool `json:"exact"`
}
