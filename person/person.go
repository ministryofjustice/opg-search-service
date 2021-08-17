package person

import (
	"encoding/json"
	"opg-search-service/response"
)

const personIndexName = "person"

type Person struct {
	ID              *int64 `json:"id"`
	UID             string `json:"uId"`
	Normalizeduid   int64  `json:"normalizedUid"`
	SageID          string `json:"sageId"`
	CaseRecNumber   string `json:"caseRecNumber"`
	Workphonenumber struct {
		ID          int    `json:"id"`
		Phonenumber string `json:"phoneNumber"`
		Type        string `json:"type"`
		Default     bool   `json:"default"`
		Classname   string `json:"className"`
	} `json:"workPhoneNumber"`
	Homephonenumber struct {
		ID          int    `json:"id"`
		Phonenumber string `json:"phoneNumber"`
		Type        string `json:"type"`
		Default     bool   `json:"default"`
		Classname   string `json:"className"`
	} `json:"homePhoneNumber"`
	Mobilephonenumber struct {
		ID          int    `json:"id"`
		Phonenumber string `json:"phoneNumber"`
		Type        string `json:"type"`
		Default     bool   `json:"default"`
		Classname   string `json:"className"`
	} `json:"mobilePhoneNumber"`
	Email             string          `json:"email"`
	Dob               string          `json:"dob"`
	Firstname         string          `json:"firstname"`
	Middlenames       string          `json:"middlenames"`
	Surname           string          `json:"surname"`
	CompanyName       string          `json:"companyName"`
	Addressline1      string          `json:"addressLine1"`
	Addressline2      string          `json:"addressLine2"`
	Addressline3      string          `json:"addressLine3"`
	Town              string          `json:"town"`
	County            string          `json:"county"`
	Postcode          string          `json:"postcode"`
	Country           string          `json:"country"`
	Isairmailrequired bool            `json:"isAirmailRequired"`
	Addresses         []PersonAddress `json:"addresses"`
	Phonenumber       string          `json:"phoneNumber"`
	Phonenumbers      []struct {
		ID          int    `json:"id"`
		Phonenumber string `json:"phoneNumber"`
		Type        string `json:"type"`
		Default     bool   `json:"default"`
		Classname   string `json:"className"`
	} `json:"phoneNumbers"`
	Persontype string `json:"personType"`
	Cases      []struct {
		UID           string `json:"uId"`
		Normalizeduid int64  `json:"normalizedUid"`
		Caserecnumber string `json:"caseRecNumber"`
		Batchid       string `json:"batchId"`
		Classname     string `json:"className"`
		Casetype      string `json:"caseType"`
		Casesubtype   string `json:"caseSubtype"`
	} `json:"cases"`
	Orders []struct {
		Order struct {
			UID           string `json:"uId"`
			Normalizeduid int64  `json:"normalizedUid"`
			Caserecnumber string `json:"caseRecNumber"`
			Batchid       string `json:"batchId"`
			Classname     string `json:"className"`
		} `json:"order"`
		Classname string `json:"className"`
	} `json:"orders"`
	Classname string `json:"className"`
}

type PersonAddress struct {
	Addresslines []string `json:"addressLines"`
	Postcode     string   `json:"postcode"`
	Classname    string   `json:"className"`
}

func (p Person) Id() int64 {
	val := int64(0)
	if p.ID != nil {
		val = *p.ID
	}
	return val
}

func (p Person) IndexName() string {
	return personIndexName
}

func (p Person) Json() string {
	b, _ := json.Marshal(p)
	return string(b)
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

func (p Person) IndexConfig() map[string]interface{} {
	return map[string]interface{}{
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
					"boost":    4.0,
				},
				"email": map[string]interface{}{
					"type":  "text",
					"boost": 4.0,
				},
				"firstname": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
					"boost":   4.0,
				},
				"middlenames": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
					"boost":   4.0,
				},
				"surname": map[string]interface{}{
					"type":    "keyword",
					"copy_to": "searchable",
					"boost":   4.0,
				},
				"companyName": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
					"boost":   4.0,
				},
				"className": map[string]interface{}{
					"type":    "text",
					"copy_to": "searchable",
					"boost":   4.0,
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
							"boost":   4.0,
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
			},
		},
	}
}
