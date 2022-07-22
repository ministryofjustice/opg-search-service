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
	keywordField := map[string]interface{}{"type": "keyword"}

	personConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"refresh_interval":   "1s",
			"analysis": map[string]interface{}{
				"filter": map[string]interface{}{
					"whitespace_remove": map[string]interface{}{
						"pattern":     " ",
						"type":        "pattern_replace",
						"replacement": "",
					},
				},
				"analyzer": map[string]interface{}{
					"default": map[string]interface{}{
						"tokenizer": "whitespace",
						"filter": []string{
							"asciifolding",
							"lowercase",
						},
					},
					"no_space_analyzer": map[string]interface{}{
						"tokenizer": "keyword",
						"filter": []string{
							"lowercase",
							"whitespace_remove",
						},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"uId":           keywordField,
				"normalizedUid": keywordField,
				"caseRecNumber": keywordField,
				"deputyNumber":  keywordField,
				"personType":    keywordField,
				"dob":           textField,
				"email":         textField,
				"firstname":     textField,
				"middlenames":   textField,
				"surname":       textField,
				"companyName":   textField,
				"className":     textField,
				"phoneNumbers": map[string]interface{}{
					"properties": map[string]interface{}{
						"phoneNumber": keywordField,
					},
				},
				"addresses": map[string]interface{}{
					"properties": map[string]interface{}{
						"addressLines": textField,
						"postcode": map[string]interface{}{
							"type": "text",
							"fields": map[string]interface{}{
								"keyword": keywordField,
							},
							"analyzer": "no_space_analyzer",
						},
					},
				},
				"cases": map[string]interface{}{
					"properties": map[string]interface{}{
						"uId":           keywordField,
						"normalizedUid": keywordField,
						"caseRecNumber": keywordField,
						"onlineLpaId":   keywordField,
						"batchId":       keywordField,
						"caseType":      keywordField,
						"caseSubtype":   keywordField,
					},
				},
				"organisationName": textField,
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
