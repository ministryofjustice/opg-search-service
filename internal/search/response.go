package search

import "encoding/json"

type Response struct {
	Results      []json.RawMessage         `json:"results"`
	Aggregations map[string]map[string]int `json:"aggregations,omitempty"`
	Total        ResponseTotal             `json:"total"`
}

type ResponseTotal struct {
	Count int  `json:"count"`
	Exact bool `json:"exact"`
}
