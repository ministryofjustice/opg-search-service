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
				"type":   "most_fields",
				"query":  "apples",
				"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname^3", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
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

func TestPrepareQueryForPersonWithPostcode(t *testing.T) {
	req := &Request{
		Term: "apples ng15ab",
		From: 123,
	}

	body := PrepareQueryForPerson(req)

	assert.Equal(t, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"multi_match": map[string]interface{}{
						"type":   "most_fields",
						"query":  "apples ng15ab",
						"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname^3", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
					}},
					{"match": map[string]interface{}{
						"addresses.postcode": map[string]interface{}{"query": "ng15ab"},
					}},
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
			"multi_match": map[string]interface{}{
				"type":   "most_fields",
				"query":  "apples",
				"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname^3", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"},
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
				"type":   "most_fields",
				"query":  "apples",
				"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname^3", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName", "firmName", "firmNumber"},
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
				"type":   "most_fields",
				"query":  "apples",
				"fields": []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname^3", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName", "firmName", "firmNumber"},
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

func TestPostcodeTerm(t *testing.T) {
	testcases := map[string]struct {
		term             string
		expectedPostcode string
	}{
		"no postcode": {
			term: "26 some street",
		},
		"short postcode": {
			term:             "26 some street s11aa john",
			expectedPostcode: "s11aa",
		},
		"long postcode": {
			term:             "26 some street ng11aa john",
			expectedPostcode: "ng11aa",
		},
		"short postcode with space": {
			term:             "26 some street s1 1aa john",
			expectedPostcode: "s1 1aa",
		},
		"long postcode with space": {
			term:             "26 some street ng1 1aa john",
			expectedPostcode: "ng1 1aa",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedPostcode, postcodeTerm(tc.term))
		})
	}
}
