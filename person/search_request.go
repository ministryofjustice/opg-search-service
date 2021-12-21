package person

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type searchRequest struct {
	Term        string   `json:"term"`
	Size        int      `json:"size,omitempty"`
	From        int      `json:"from"`
	PersonTypes []string `json:"person_types"`
}

func CreateSearchRequestFromRequest(r *http.Request) (*searchRequest, error) {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	if buf.Len() == 0 {
		return nil, errors.New("request body is empty")
	}

	var req searchRequest
	err := json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to unmarshal JSON request")
	}

	req.sanitise()

	if req.Term == "" {
		return nil, errors.New("search term is required and cannot be empty")
	}

	return &req, nil
}

func (sr *searchRequest) sanitise() {
	re := regexp.MustCompile(`[^’'\p{L}\d\-.@ \/_]`)
	log.Println(re.ReplaceAllString(sr.Term, ""))
	sr.Term = strings.TrimSpace(re.ReplaceAllString(sr.Term, ""))

	for i, val := range sr.PersonTypes {
		sr.PersonTypes[i] = strings.TrimSpace(re.ReplaceAllString(val, ""))
	}
}
