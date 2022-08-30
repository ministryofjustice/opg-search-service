package deputy

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasName = "deputy"

type Deputy struct {
	ID               *int64 `json:"id"`
	UID              string `json:"uId"`
	Normalizeduid    int64  `json:"normalizedUid"`
	DeputyNumber     *int64 `json:"deputyNumber"`
	Persontype       string `json:"personType"`
	Dob              string `json:"dob"`
	Firstname        string `json:"firstname"`
	Middlenames      string `json:"middlenames"`
	Surname          string `json:"surname"`
	Othernames       string `json:"othernames"`
	CompanyName      string `json:"companyName"`
	Classname        string `json:"className"`
	OrganisationName string `json:"organisationName"`
}

func (p Deputy) Id() int64 {
	val := int64(0)
	if p.ID != nil {
		val = *p.ID
	}
	return val
}

func (p Deputy) Validate() []response.Error {
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

	deputyConfig := map[string]interface{}{
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
				"searchable":       textField,
				"uId":              searchableKeywordField,
				"normalizedUid":    searchableKeywordField,
				"deputyNumber":     searchableKeywordField,
				"personType":       keywordField,
				"dob":              searchableTextField,
				"firstname":        searchableTextField,
				"middlenames":      searchableTextField,
				"surname":          searchableTextField,
				"othernames":       searchableTextField,
				"companyName":      searchableTextField,
				"className":        searchableTextField,
				"organisationName": searchableTextField,
			},
		},
	}

	data, err := json.Marshal(deputyConfig)
	if err != nil {
		return "", nil, err
	}

	sum := sha256.Sum256(data)

	return fmt.Sprintf("%s_%x", AliasName, sum[:8]), data, err
}
