package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareQueryForFirm(t *testing.T) {
	req := &Request{
		Term: "apples",
		From: 123,
	}

	body := PrepareQueryForFirm(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]string{
					"field": "personType",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)
}

func TestPrepareQueryForFirmWithOptions(t *testing.T) {
	req := &Request{
		Term:        "apples",
		From:        123,
		Size:        10,
		PersonTypes: []string{"deputy", "donor"},
	}

	body := PrepareQueryForFirm(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]string{
					"field": "personType",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{
			map[string]interface{}{"term": map[string]string{"personType": "deputy"}},
			map[string]interface{}{"term": map[string]string{"personType": "donor"}},
		}}},
		"from": 123,
		"size": 10,
	}, body)
}

func TestPrepareQueryForPerson(t *testing.T) {
	req := &Request{
		Term: "apples",
		From: 123,
	}

	body := PrepareQueryForPerson(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": "apples",
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
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)
}

func TestPrepareQueryForPersonWithOptions(t *testing.T) {
	req := &Request{
		Term:        "apples",
		From:        123,
		Size:        10,
		PersonTypes: []string{"deputy", "donor"},
	}

	body := PrepareQueryForPerson(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": "apples",
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
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{
			map[string]interface{}{"term": map[string]string{"personType": "deputy"}},
			map[string]interface{}{"term": map[string]string{"personType": "donor"}},
		}}},
		"from": 123,
		"size": 10,
	}, body)
}

func TestPrepareQueryForFirmAndPerson(t *testing.T) {
	req := &Request{
		Term: "apples",
		From: 123,
	}

	body := PrepareQueryForFirmAndPerson(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]string{
					"field": "personType",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)
}

func TestPrepareQueryForDeputy(t *testing.T) {
	req := &Request{
		Term: "Niko",
		From: 9,
	}

	body := PrepareQueryForDeputy(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": "Niko",
						"fields": []string{
							"firstname", "othernames", "middlenames", "surname", "organisationName",
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
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        9,
	}, body)
}

func TestPrepareQueryForFirmAndPersonWithOptions(t *testing.T) {
	req := &Request{
		Term:        "apples",
		From:        123,
		Size:        10,
		PersonTypes: []string{"deputy", "donor"},
	}

	body := PrepareQueryForFirmAndPerson(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]string{
					"field": "personType",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{
			map[string]interface{}{"term": map[string]string{"personType": "deputy"}},
			map[string]interface{}{"term": map[string]string{"personType": "donor"}},
		}}},
		"from": 123,
		"size": 10,
	}, body)
}
