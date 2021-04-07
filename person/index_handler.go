package person

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/response"
	"time"
)

type IndexHandler struct {
	logger *log.Logger
	es     elasticsearch.ClientInterface
}

func NewIndexHandler(logger *log.Logger) (*IndexHandler, error) {
	client, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Println(err)
		return nil, errors.New("unable to create a new Elasticsearch client")
	}

	return &IndexHandler{
		logger,
		client,
	}, nil
}

func (i IndexHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// get persons from payload
	var req IndexRequest
	resp := new(response.IndexResponse)

	bodyBuf := new(bytes.Buffer)
	_, _ = bodyBuf.ReadFrom(r.Body)

	if bodyBuf.Len() == 0 {
		i.logger.Println("request body is empty")
		response.WriteJSONError(rw, "request", "Request body is empty", http.StatusBadRequest)
		return
	}

	err := json.Unmarshal(bodyBuf.Bytes(), &req)
	if err != nil {
		i.logger.Println(err.Error())
		response.WriteJSONError(rw, "request", "Unable to unmarshal JSON request", http.StatusBadRequest)
		return
	}

	validationErrs := req.Validate()
	if len(validationErrs) > 0 {
		i.logger.Println("Request failed validation", validationErrs)
		response.WriteJSONErrors(rw, "Some fields have failed validation", validationErrs, http.StatusBadRequest)
		return
	}

	for _, p := range req.Persons {
		// index person in elasticsearch
		resp.Results = append(resp.Results, *i.es.Index(p))
	}

	jsonResp, _ := json.Marshal(resp)

	rw.WriteHeader(http.StatusAccepted)

	_, _ = rw.Write(jsonResp)

	i.logger.Println("Request took: ", time.Since(start))
}
