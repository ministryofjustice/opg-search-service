package Merged

import (
"crypto/sha256"
"encoding/json"
"fmt"

"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasNamePerson = "person"
const AliasNameFirm = "firm"

type Person struct {
	EntityId
	UID              string              `json:"uId"`
	Normalizeduid    int64               `json:"normalizedUid"`
	CaseRecNumber    string              `json:"caseRecNumber"`
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

type Firm struct {
	EntityId
	Persontype   string `json:"personType"`
	Email        string `json:"email"`
	FirmName     string `json:"firmName"`
	FirmNumber   int  	`json:"firmNumber"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	AddressLine3 string `json:"addressLine3"`
	Town         string `json:"town"`
	County       string `json:"county"`
	Postcode     string `json:"postcode"`
	Phonenumber  string `json:"phoneNumber"`
}

type EntityId struct {
	ID *int64 `json:"id"`
}

type EntityIdInterface interface {
	Id()
}

func(e EntityId) Id() int64 {
	val := int64(0)
	if e.ID != nil {
		val = *e.ID
	}
	return val
}

func (e EntityId) Validate() []response.Error {
	var errs []response.Error

	if e.ID == nil {
		errs = append(errs, response.Error{
			Name:        "id",
			Description: "field is empty",
		})
	}

	return errs
}

func EntityIndexConfig() (pName, fName string, pConfig []byte, fConfig []byte, err error) {
	personConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"refresh_interval":   "1s",
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"quick_search": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "whitespace",
						"filter": []string{
							"asciifolding",
							"lowercase",
						},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"searchable": map[string]interface{}{
					"type":     "text",
					"analyzer": "quick_search",
				},
				"uId": map[string]interface{}{
					"type":    "keyword",
					"copy_to": "searchable",
				},
				"normalizedUid": map[string]interface{}{
					"type":    "text",
					"index":   false,
					"copy_to": "searchable",
				},
				"caseRecNumber": map[string]interface{}{
					"type":    "keyword",
					"copy_to": "searchable",
				},
				"personType": map[string]interface{}{
					"type": "keyword",
				},
				"dob": map[string]interface{}{
					"type":     "text",
					"analyzer": "whitespace",
					"copy_to":  "searchable",
				},
				"email": map[string]interface{}{
					"type":  "text",
				},
				"firstname": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
				},
				"middlenames": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
				},
				"surname": map[string]interface{}{
					"type":    "keyword",
					"copy_to": "searchable",
				},
				"companyName": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
				},
				"className": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
				},
				"phoneNumbers": map[string]interface{}{
					"properties": map[string]interface{}{
						"phoneNumber": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
					},
				},
				"addresses": map[string]interface{}{
					"properties": map[string]interface{}{
						"addressLines": map[string]interface{}{
							"type":    "text",
							"copy_to": "searchable",
						},
						"postcode": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
					},
				},
				"cases": map[string]interface{}{
					"properties": map[string]interface{}{
						"uId": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
						"normalizedUid": map[string]interface{}{
							"type":    "text",
							"index":   false,
							"copy_to": "searchable",
						},
						"caseRecNumber": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
						"onlineLpaId": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
						"batchId": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
						"caseType": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
						"caseSubtype": map[string]interface{}{
							"type":    "keyword",
							"copy_to": "searchable",
						},
					},
				},
				"organisationName": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
				},
			},
		},
	}

	firmConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
			"refresh_interval":   "1s",
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"personType": map[string]interface{}{
					"type": "keyword",
				},
				"email": map[string]interface{}{
					"type": "text",
				},
				"firmName": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"raw": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"firmNumber": map[string]interface{}{
					"type": "keyword",
				},
				"phoneNumber": map[string]interface{}{
					"type": "keyword",
				},
				"addressLine1": map[string]interface{}{
					"type": "text",
				},
				"addressLine2": map[string]interface{}{
					"type": "text",
				},
				"addressLine3": map[string]interface{}{
					"type": "text",
				},
				"town": map[string]interface{}{
					"type": "text",
				},
				"county": map[string]interface{}{
					"type": "text",
				},
				"postcode": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	personData, err := json.Marshal(personConfig)
	firmData, err := json.Marshal(firmConfig)
	if err != nil {
		return "", "", nil, nil, err
	}

	sumPersonData := sha256.Sum256(personData)
	sumFirmData := sha256.Sum256(firmData)

	return fmt.Sprintf("%s_%x", AliasNamePerson, sumPersonData[:8]), fmt.Sprintf("%s_%x", AliasNameFirm, sumFirmData[:8]), personData, firmData, err
}