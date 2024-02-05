package digitallpa

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasName = "digitalLpa"

type Person struct {
	Firstnames string  `json:"firstNames"`
	Surname    string  `json:"surname"`
	Address    Address `json:"address"`
}

type Donor struct {
	Person
	Dob string `json:"dob"`
}

type Attorney struct {
	Person
	Dob string `json:"dob"`
}

type Address struct {
	Line1    string `json:"line1"`
	Line2    string `json:"line2"`
	Line3    string `json:"line3"`
	Postcode string `json:"postcode"`
}

type DigitalLpa struct {
	Uid                 string     `json:"uId"`
	LpaType             string     `json:"lpaType"`
	Donor               Donor      `json:"donor"`
	CertificateProvider Person     `json:"certificateProvider"`
	Attorneys           []Attorney `json:"attorneys"`
}

func (d DigitalLpa) Id() string {
	return d.Uid
}

func (d DigitalLpa) Validate() []response.Error {
	var errs []response.Error

	if len(d.Uid) == 0 {
		errs = append(errs, response.Error{
			Name:        "uId",
			Description: "field is empty",
		})
	}

	return errs
}

func IndexConfig() (name string, config []byte, err error) {
	textField := map[string]interface{}{"type": "text"}
	searchableTextField := map[string]interface{}{"type": "text", "copy_to": "searchable"}
	keywordField := map[string]interface{}{"type": "keyword"}

	digitalLpaConfig := map[string]interface{}{
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
				"searchable": textField,
				"uid":        searchableTextField,
				"lpaType":    textField,
				"donor": map[string]interface{}{
					"properties": map[string]interface{}{
						"firstnames": searchableTextField,
						"surname":    searchableTextField,
						"dob":        searchableTextField,
						"address": map[string]interface{}{
							"properties": map[string]interface{}{
								"line1": searchableTextField,
								"line2": searchableTextField,
								"line3": searchableTextField,
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
					},
				},
				"certificateProvider": map[string]interface{}{
					"properties": map[string]interface{}{
						"firstnames": searchableTextField,
						"surname":    searchableTextField,
						"address": map[string]interface{}{
							"properties": map[string]interface{}{
								"line1": searchableTextField,
								"line2": searchableTextField,
								"line3": searchableTextField,
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
					},
				},
				"attorneys": map[string]interface{}{
					"properties": map[string]interface{}{
						"firstnames": searchableTextField,
						"surname":    searchableTextField,
						"dob":        searchableTextField,
						"address": map[string]interface{}{
							"properties": map[string]interface{}{
								"line1": searchableTextField,
								"line2": searchableTextField,
								"line3": searchableTextField,
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
					},
				},
			},
		},
	}

	data, err := json.Marshal(digitalLpaConfig)
	if err != nil {
		return "", nil, err
	}

	sum := sha256.Sum256(data)

	return fmt.Sprintf("%s_%x", AliasName, sum[:8]), data, err
}
