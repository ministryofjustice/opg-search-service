package person

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasName = "person"

type Person struct {
	ID               *int64              `json:"id"`
	UID              string              `json:"uId"`
	Normalizeduid    int64               `json:"normalizedUid"`
	CaseRecNumber    string              `json:"caseRecNumber"`
	DeputyNumber     *int64              `json:"deputyNumber"`
	Persontype       string              `json:"personType"`
	Dob              string              `json:"dob"`
	Email            string              `json:"email"`
	Firstname        string              `json:"firstname"`
	Middlenames      string              `json:"middlenames"`
	Surname          string              `json:"surname"`
	Previousnames    string              `json:"previousnames"`
	Othernames       string              `json:"othernames"`
	CompanyName      string              `json:"companyName"`
	Classname        string              `json:"className"`
	OrganisationName string              `json:"organisationName"`
	Addresses        []PersonAddress     `json:"addresses"`
	Phonenumbers     []PersonPhonenumber `json:"phoneNumbers"`
	Cases            []PersonCase        `json:"cases"`
}

type PersonCase struct {
	UID           string `json:"uId"`
	Normalizeduid int64  `json:"normalizedUid"`
	Caserecnumber string `json:"caseRecNumber"`
	OnlineLpaId   string `json:"onlineLpaId"`
	Batchid       string `json:"batchId"`
	Casetype      string `json:"caseType"`
	Casesubtype   string `json:"caseSubtype"`
}

type PersonAddress struct {
	Addresslines []string `json:"addressLines"`
	Postcode     string   `json:"postcode"`
}

type PersonPhonenumber struct {
	Phonenumber string `json:"phoneNumber"`
}

func (p Person) Id() int64 {
	val := int64(0)
	if p.ID != nil {
		val = *p.ID
	}
	return val
}

func (p Person) Validate() []response.Error {
	var errs []response.Error

	if p.ID == nil {
		errs = append(errs, response.Error{
			Name:        "id",
			Description: "field is empty",
		})
	}

	return errs
}

func IndexConfig() (name string, config []byte, err error) {
	textField := map[string]interface{}{"type": "text"}
	searchableTextField := map[string]interface{}{"type": "text", "copy_to": "searchable"}
	keywordField := map[string]interface{}{"type": "keyword"}
	searchableKeywordField := map[string]interface{}{"type": "keyword", "copy_to": "searchable"}
	//composite of first and surname
	fullNameSearchableText := map[string]interface{}{"type": "text", "copy_to": []string{"full_name", "searchable"}}

	personConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"refresh_interval":   "1s",
			"analysis": map[string]interface{}{
				"filter": map[string]interface{}{
					"whitespace_remove": map[string]interface{}{
						"type":        "pattern_replace",
						"pattern":     " ",
						"replacement": "",
					},
				},
				"analyzer": map[string]interface{}{
					"default": map[string]interface{}{
						"tokenizer": "whitespace",
						"filter":    []string{"asciifolding", "lowercase"},
					},
					"no_space_analyzer": map[string]interface{}{
						"tokenizer": "keyword",
						"filter":    []string{"whitespace_remove", "lowercase"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"searchable":    textField,
				"full_name":     textField,
				"uId":           searchableKeywordField,
				"normalizedUid": searchableKeywordField,
				"caseRecNumber": searchableKeywordField,
				"deputyNumber":  searchableKeywordField,
				"personType":    keywordField,
				"dob":           searchableTextField,
				"email":         textField,
				"firstname":     fullNameSearchableText,
				"middlenames":   searchableTextField,
				"surname":       fullNameSearchableText,
				"previousnames": searchableTextField,
				"othernames":    searchableTextField,
				"companyName":   searchableTextField,
				"className":     searchableTextField,
				"clientSource": map[string]interface{}{
					"type":  "text",
					"index": false,
				},
				"phoneNumbers": map[string]interface{}{
					"properties": map[string]interface{}{
						"phoneNumber": searchableKeywordField,
					},
				},
				"addresses": map[string]interface{}{
					"properties": map[string]interface{}{
						"addressLines": searchableTextField,
						"postcode": map[string]interface{}{
							"type":     "text",
							"analyzer": "no_space_analyzer",
							"copy_to":  "searchable",
							"fields": map[string]interface{}{
								"keyword": keywordField,
							},
						},
					},
				},
				"cases": map[string]interface{}{
					"properties": map[string]interface{}{
						"uId":           searchableKeywordField,
						"normalizedUid": searchableTextField,
						"caseRecNumber": searchableKeywordField,
						"onlineLpaId":   searchableKeywordField,
						"batchId":       searchableKeywordField,
						"caseType":      searchableKeywordField,
						"caseSubtype":   searchableKeywordField,
					},
				},
				"organisationName": searchableTextField,
			},
		},
	}

	data, err := json.Marshal(personConfig)
	if err != nil {
		return "", nil, err
	}

	sum := sha256.Sum256(data)

	return fmt.Sprintf("%s_%x", AliasName, sum[:8]), data, err
}
