package poadraftapplication

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasName = "poadraftapplication"

type DraftApplication struct {
	ID *int64 `json:"id"`
	DonorName string `json:"donorName"`
	DonorEmail string `json:"donorEmail"`
	DonorPhone string `json:"donorPhone"`
	DonorAddressLine1 string `json:"donorAddressLine1"`
	DonorPostcode string `json:"donorPostcode"`
	CorrespondentName string `json:"correspondentName"`
	CorrespondentAddressLine1 string `json:"correspondentAddressLine1"`
	CorrespondentPostcode string `json:"correspondentPostcode"`
}

func (f DraftApplication) Id() int64 {
	val := int64(0)
	if f.ID != nil {
		val = *f.ID
	}
	return val
}

func (f DraftApplication) Validate() []response.Error {
	var errs []response.Error

	if f.ID == nil {
		errs = append(errs, response.Error{
			Name: "id",
			Description: "field is empty",
		})
	}

	return errs
}

func IndexConfig() (name string, config []byte, err error) {
	textField := map[string]interface{}{"type": "text"}
	searchableTextField := map[string]interface{}{"type": "text", "copy_to": "draftApplicationSearchable"}

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
				"draftApplicationSearchable": textField,
				"donorName": searchableTextField,
				"donorEmail": searchableTextField,
				"donorPhone": searchableTextField,
				"donorAddressLine1": searchableTextField,
				"donorPostcode": map[string]interface{}{
					"type": "text",
					"analyzer": "no_space_analyzer",
					"copy_to": "draftApplicationSearchable",
				},
				"correspondentName": searchableTextField,
				"correspondentAddressLine1": searchableTextField,
				"correspondentPostcode": map[string]interface{}{
					"type": "text",
					"analyzer": "no_space_analyzer",
					"copy_to": "draftApplicationSearchable",
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
