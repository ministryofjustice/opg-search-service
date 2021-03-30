package person

import "encoding/json"

type Person struct {
	UID             string `json:"uId"`
	Normalizeduid   int64  `json:"normalizedUid"`
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
	Email             string `json:"email"`
	Dob               string `json:"dob"`
	Firstname         string `json:"firstname"`
	Middlenames       string `json:"middlenames"`
	Surname           string `json:"surname"`
	Addressline1      string `json:"addressLine1"`
	Addressline2      string `json:"addressLine2"`
	Addressline3      string `json:"addressLine3"`
	Town              string `json:"town"`
	County            string `json:"county"`
	Postcode          string `json:"postcode"`
	Country           string `json:"country"`
	Isairmailrequired bool   `json:"isAirmailRequired"`
	Addresses         []struct {
		Addresslines []string `json:"addressLines"`
		Postcode     string   `json:"postcode"`
		Classname    string   `json:"className"`
	} `json:"addresses"`
	Phonenumber  string `json:"phoneNumber"`
	Phonenumbers []struct {
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

func (p Person) Id() int64 {
	return p.Normalizeduid
}

func (p Person) IndexName() string {
	return "person"
}

func (p Person) Json() string {
	b, _ := json.Marshal(p)
	return string(b)
}
