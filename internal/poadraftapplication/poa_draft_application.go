package poadraftapplication

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasName = "poadraftapplication"

type DraftApplicationDonor struct {
	Name string `json:"name"`
	Dob string `json:"dob"`
	Postcode string `json:"postcode"`
}

type DraftApplication struct {
	UID *string `json:"uId"`
	Donor DraftApplicationDonor `json:"donor"`
}

func (d DraftApplication) Id() string {
	if d.UID == nil {
		return "0"
	}
	return *d.UID
}

func (d DraftApplication) Validate() []response.Error {
	var errs []response.Error

	if d.UID == nil || len(*d.UID) == 0 {
		errs = append(errs, response.Error{
			Name: "uId",
			Description: "field is empty",
		})
	}

	return errs
}

func IndexConfig() (name string, config []byte, err error) {
	textField := map[string]interface{}{"type": "text"}
	searchableTextField := map[string]interface{}{"type": "text", "copy_to": "searchable"}

	draftApplicationConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards": 1,
			"number_of_replicas": 1,
			"refresh_interval": "1s",
			"analysis": map[string]interface{}{
				"filter": map[string]interface{}{
					"whitespace_remove": map[string]interface{}{
						"type": "pattern_replace",
						"pattern": " ",
						"replacement": "",
					},
				},
				"analyzer": map[string]interface{}{
					"default": map[string]interface{}{
						"tokenizer": "whitespace",
						"filter": []string{"asciifolding", "lowercase"},
					},
					"no_space_analyzer": map[string]interface{}{
						"tokenizer": "keyword",
						"filter": []string{"whitespace_remove", "lowercase"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"searchable": textField,
				"uId": searchableTextField,
				"donor": map[string]interface{}{
					"properties": map[string]interface{}{
						"name": searchableTextField,
						"dob": searchableTextField,
						"postcode": map[string]interface{}{
							"type": "text",
							"analyzer": "no_space_analyzer",
							"copy_to": "searchable",
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(draftApplicationConfig)
	if err != nil {
		return "", nil, err
	}

	sum := sha256.Sum256(data)

	return fmt.Sprintf("%s_%x", AliasName, sum[:8]), data, err
}
