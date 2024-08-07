package firm

import (
	"encoding/json"
	"strconv"

	"github.com/ministryofjustice/opg-search-service/internal/response"
)

const AliasName = "firm"

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

func (f Firm) Id() string {
	val := 0
	if f.ID != nil {
		val = int(*f.ID)
	}
	return strconv.Itoa(val)
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

func IndexConfig() (config []byte, err error) {
	firmConfig := map[string]interface{}{
		"settings": map[string]interface{}{
			"number_of_shards":   3,
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
		return nil, err
	}

	return data, err
}
