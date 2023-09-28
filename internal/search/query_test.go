package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/ministryofjustice/opg-search-service/internal/poadraftapplication"
)

func TestPrepareQueryForDraftApplication(t *testing.T) {
	req := &Request{
		Term: "MMMoooossssa",
		From: 123,
	}

	indices, body := PrepareQueryForDraftApplication(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": "MMMoooossssa",
						"fields": []string{
							"searchable",
						},
						"default_operator": "AND",
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)

	assert.Equal(t, []string{poadraftapplication.AliasName}, indices)
}

func TestPrepareQueryForFirm(t *testing.T) {
	req := &Request{
		Term: "apples",
		From: 123,
	}

	indices, body := PrepareQueryForFirm(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)

	assert.Equal(t, []string{firm.AliasName}, indices)
}

func TestPrepareQueryForFirmWithOptions(t *testing.T) {
	req := &Request{
		Term:        "apples",
		From:        123,
		Size:        10,
		PersonTypes: []string{"deputy", "donor"},
	}

	indices, body := PrepareQueryForFirm(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
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

	assert.Equal(t, []string{firm.AliasName}, indices)
}

func TestPrepareQueryForPerson(t *testing.T) {
	req := &Request{
		Term: "apples",
		From: 123,
	}

	indices, body := PrepareQueryForPerson(req)

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
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)

	assert.Equal(t, []string{person.AliasName}, indices)
}

func TestPrepareQueryForPersonWithOptions(t *testing.T) {
	req := &Request{
		Term:        "apples",
		From:        123,
		Size:        10,
		PersonTypes: []string{"deputy", "donor"},
	}

	indices, body := PrepareQueryForPerson(req)

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
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
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

	assert.Equal(t, []string{person.AliasName}, indices)
}

func TestPrepareQueryForPersonAlreadyPrepared(t *testing.T) {
	req := &Request{
		Prepared: map[string]interface{}{
			"query": "some prepared query",
		},
	}

	indices, body := PrepareQueryForPerson(req)

	assert.Equal(t, map[string]interface{}{
		"query": "some prepared query",
	}, body)

	assert.Equal(t, []string{person.AliasName}, indices)
}

func TestPrepareQueryForDeputy(t *testing.T) {
	req := &Request{
		Term: "Niko",
		From: 9,
	}

	indices, body := PrepareQueryForDeputy(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"simple_query_string": map[string]interface{}{
						"query": "Niko",
						"fields": []string{
							"firstname", "middlenames", "surname", "previousnames", "othernames", "organisationName",
						},
						"default_operator": "AND",
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        9,
	}, body)

	assert.Equal(t, []string{person.AliasName}, indices)
}

func TestPrepareQueryForAll(t *testing.T) {
	req := &Request{
		Term:    "apples",
		From:    123,
	}

	indices, body := PrepareQueryForAll(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)

	assert.Equal(t, []string{firm.AliasName, person.AliasName, poadraftapplication.AliasName}, indices)
}

func TestPrepareQueryForAllEmptyIndices(t *testing.T) {
	req := &Request{
		Term:    "apples",
		From:    123,
		Indices: []string{},
	}

	indices, body := PrepareQueryForAll(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
				},
			},
		},
		"post_filter": map[string]interface{}{"bool": map[string]interface{}{"should": []interface{}{}}},
		"from":        123,
	}, body)

	assert.Equal(t, []string{firm.AliasName, person.AliasName, poadraftapplication.AliasName}, indices)
}

func TestPrepareQueryForAllWithOptions(t *testing.T) {
	req := &Request{
		Term:        "apples",
		From:        123,
		Size:        10,
		PersonTypes: []string{"deputy", "donor"},
		Indices:     []string{"firm", "person"},
	}

	indices, body := PrepareQueryForAll(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"firmName", "firmNumber", "caseRecNumber", "searchable"},
			},
		},
		"aggs": map[string]interface{}{
			"personType": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "personType",
					"size":  "20",
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

	assert.Equal(t, []string{firm.AliasName, person.AliasName}, indices)
}
