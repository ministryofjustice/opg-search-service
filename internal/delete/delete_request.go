package delete

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Request struct {
	Uid string `json:"uid"`
}

func parseDeleteRequest(r *http.Request) (*Request, error) {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	if buf.Len() == 0 {
		return nil, errors.New("request body is empty")
	}

	var req Request
	err := json.Unmarshal(buf.Bytes(), &req)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to unmarshal JSON request")
	}

	if req.Uid == "" {
		return nil, errors.New("uid is required and cannot be empty")
	}

	return &req, nil
}
