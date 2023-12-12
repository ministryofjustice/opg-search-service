package search

import (
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
)

var firmIndices = []string{firm.AliasName}
var personIndices = []string{person.AliasName}
var allIndices = []string{firm.AliasName, person.AliasName}

func PrepareQueryForFirm(req *Request) ([]string, map[string]interface{}) {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Term,
				"fields": []string{"firmName", "firmNumber"},
			},
		},
	}

	return firmIndices, withDefaults(req, body)
}

func PrepareQueryForPerson(req *Request) ([]string, map[string]interface{}) {
	if req.Prepared != nil {
		return personIndices, req.Prepared
	}

	body := map[string]interface{}{
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
	}

	return personIndices, withDefaults(req, body)
}

func PrepareQueryForDeputy(req *Request) ([]string, map[string]interface{}) {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": req.Term,
						"fields": []string{
							"firstname",
							"middlenames",
							"surname",
							"previousnames",
							"othernames",
							"organisationName",
						},
						"default_operator": "AND",
					},
				},
			},
		},
	}

	return personIndices, withDefaults(req, body)
}

func PrepareQueryForAll(req *Request) ([]string, map[string]interface{}) {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Term,
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
	}

	indices := allIndices

	if len(req.Indices) > 0 {
		indices = req.Indices
	}

	return indices, withDefaults(req, body)
}

func withDefaults(req *Request, body map[string]interface{}) map[string]interface{} {
	// initialised as elasticsearch will error with nil
	filters := []interface{}{}

	for _, f := range req.PersonTypes {
		filters = append(filters, map[string]interface{}{
			"term": map[string]string{
				"personType": f,
			},
		})
	}

	body["aggs"] = map[string]interface{}{
		"personType": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "personType",
				"size":  "20",
			},
		},
	}
	body["from"] = req.From
	body["post_filter"] = map[string]interface{}{"bool": map[string]interface{}{"should": filters}}
	if req.Size > 0 {
		body["size"] = req.Size
	}

	return body
}
