package person

import (
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
	es     *elasticsearch.Client
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		i.logger.Println(err.Error())
		http.Error(rw, "Unable to decode JSON request", http.StatusBadRequest)
		return
	}

	for _, p := range req.Persons {
		// index person in elasticsearch
		resp.Results = append(resp.Results, *i.es.Index(p))
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		i.logger.Println(err.Error())
		http.Error(rw, "Unable to encode response object to JSON", http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusAccepted)

	_, err = rw.Write(jsonResp)
	if err != nil {
		i.logger.Println(err.Error())
	}

	i.logger.Println("Request took: ", time.Since(start))
}
