package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/person"
	"opg-search-service/request"
)

type IndexPersonHandler struct {
	logger *log.Logger
	es     *elasticsearch.Client
}

func NewIndexPersonHandler(logger *log.Logger) (*IndexPersonHandler, error) {
	client, err := elasticsearch.NewClient(logger)
	if err != nil {
		logger.Println(err)
		return nil, errors.New("unable to create a new Elasticsearch client")
	}

	return &IndexPersonHandler{
		logger,
		client,
	}, nil
}

func (i IndexPersonHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// get person IDs from payload
	var ir request.IndexRequest

	err := json.NewDecoder(r.Body).Decode(&ir)
	if err != nil {
		i.logger.Println("unable to decode JSON request", r.Body)
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	fail := false

	for _, id := range ir.Ids {
		p := person.Person{
			FirstName: fmt.Sprintf("Test %d", id),
			LastName:  "Test",
		}

		// index person in elasticsearch
		err = i.es.Index(p)
		if err != nil {
			i.logger.Println(err)
			i.logger.Println("unable to index person", id)
			fail = true
		} else {
			i.logger.Println("person indexed successfully", id)
		}
	}

	if fail {
		errString := "one or more of the indexing requests has failed"
		i.logger.Println(errString)
		http.Error(rw, errString, http.StatusInternalServerError)
	}
}
