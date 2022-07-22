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
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
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
			"multi_match": map[string]interface{}{
				"query":  "apples",
				"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
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
				"fields": []string{"firmName", "firmNumber", "uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
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
				"fields": []string{"firmName", "firmNumber", "uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
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
