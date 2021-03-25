package person

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"opg-search-service/elasticsearch"
)

type IndexHandler struct {
	logger *log.Logger
	es     *elasticsearch.Client
}

func NewIndexHandler(logger *log.Logger) (*IndexHandler, error) {
	client, err := elasticsearch.NewClient(logger)
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
	// get persons from payload
	var ir IndexRequest

	err := json.NewDecoder(r.Body).Decode(&ir)
	if err != nil {
		i.logger.Println("unable to decode JSON request", r.Body)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	fail := false

	for _, p := range ir.Persons {
		// index person in elasticsearch
		err = i.es.Index(p)
		if err != nil {
			i.logger.Println(err)
			i.logger.Println("unable to index person", p.Id)
			fail = true
		} else {
			i.logger.Println("person indexed successfully", p.Id)
		}
	}

	if fail {
		errString := "one or more of the indexing requests has failed"
		i.logger.Println(errString)
		http.Error(rw, errString, http.StatusInternalServerError)
	}

	rw.WriteHeader(http.StatusAccepted)
}
