package search

func PrepareQueryForFirm(req *Request) map[string]interface{} {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Term,
				"fields": []string{"firmName", "firmNumber"},
			},
		},
	}

	return withDefaults(req, body)
}

func PrepareQueryForPerson(req *Request) map[string]interface{} {
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

	return withDefaults(req, body)
}

func PrepareQueryForDeputy(req *Request) map[string]interface{} {
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
							"organisationName",
						},
						"default_operator": "AND",
					},
				},
			},
		},
	}

	return withDefaults(req, body)
}

func PrepareQueryForFirmAndPerson(req *Request) map[string]interface{} {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Term,
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
	}

	return withDefaults(req, body)
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
			"terms": map[string]string{
				"field": "personType",
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
