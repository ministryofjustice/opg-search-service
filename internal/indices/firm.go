package indices

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasNameFirm = "firm"
const AliasNamePersonFirm = "person,firm"

type Entity interface {
	Id () int64
	Validate() []response.Error
}

type Firm struct {
	ID           *int64 `json:"id"`
	Persontype   string `json:"personType"`
	Email        string `json:"email"`
	FirmName     string `json:"firmName"`
	FirmNumber   string `json:"firmNumber"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
	AddressLine3 string `json:"addressLine3"`
	Town         string `json:"town"`
	County       string `json:"county"`
	Postcode     string `json:"postcode"`
	Phonenumber  string `json:"phoneNumber"`
}

func (f Firm) Id() int64 {
	val := int64(0)
	if f.ID != nil {
		val = *f.ID
	}
	return val
}

func (f Firm) Validate() []response.Error {
	var errs []response.Error

	if f.ID == nil {
		errs = append(errs, response.Error{
			Name:        "id",
			Description: "field is empty",
		})
	}

	return errs
}

func IndexConfigFirm() (name string, config []byte, err error) {
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

	data, err := json.Marshal(firmConfig)
	if err != nil {
		return "", nil, err
	}

	sum := sha256.Sum256(data)

	return fmt.Sprintf("%s_%x", AliasNameFirm, sum[:8]), data, err
}
