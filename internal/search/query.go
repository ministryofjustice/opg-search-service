package search

import (
	"regexp"
	"strings"
)

var (
	firmFields   = []string{"firmName", "firmNumber"}
	personFields = []string{"uId", "normalizedUid", "caseRecNumber", "deputyNumber", "dob", "firstname", "middlenames", "surname^3", "companyName", "className", "phoneNumbers.phoneNumber", "addresses.addressLines", "addresses.postcode", "cases.uId", "cases.normalizedUid", "cases.caseRecNumber", "cases.onlineLpaId", "cases.batchId", "cases.caseType", "cases.caseSubtype", "organisationName"}
)

func PrepareQueryForFirm(req *Request) map[string]interface{} {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Term,
				"fields": firmFields,
			},
		},
	}

	return withDefaults(req, body)
}

func PrepareQueryForPerson(req *Request) map[string]interface{} {
	postcode := postcodeTerm(req.Term)

	multiMatch := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"type":   "most_fields",
			"query":  req.Term,
			"fields": personFields,
		},
	}

	if postcode == "" {
		return withDefaults(req, map[string]interface{}{
			"query": multiMatch,
		})
	}

	return withDefaults(req, map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					multiMatch,
					{
						"match": map[string]interface{}{
							"addresses.postcode": map[string]interface{}{"query": postcode},
						},
					},
				},
			},
		},
	})
}

func PrepareQueryForFirmAndPerson(req *Request) map[string]interface{} {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"type":   "most_fields",
				"query":  req.Term,
				"fields": append(personFields, firmFields...),
			},
		},
	}

	return withDefaults(req, body)
}

func postcodeTerm(term string) string {
	re, _ := regexp.Compile(`(gir 0a{2})|((([a-z][0-9]{1,2})|(([a-z][a-hj-y][0-9]{1,2})|(([a-z][0-9][a-z])|([a-z][a-hj-y][0-9][a-z]?))))\s?[0-9][a-z]{2})`)
	matches := re.FindStringSubmatch(strings.ToLower(term))

	if len(matches) > 0 {
		return matches[0]
	}

	return ""
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
