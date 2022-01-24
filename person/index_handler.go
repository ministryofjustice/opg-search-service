package person

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"opg-search-service/elasticsearch"
	"opg-search-service/response"
	"time"

	"github.com/sirupsen/logrus"
)

type IndexClient interface {
	DoBulk(op *elasticsearch.BulkOp) (elasticsearch.BulkResult, error)
}

type IndexHandler struct {
	logger    *logrus.Logger
	client    IndexClient
	indexName string
}

func NewIndexHandler(logger *logrus.Logger, indexName string) (*IndexHandler, error) {
	client, err := elasticsearch.NewClient(&http.Client{}, logger)
	if err != nil {
		logger.Println(err)
		return nil, errors.New("unable to create a new Elasticsearch client")
	}

	return &IndexHandler{
		logger:    logger,
		client:    client,
		indexName: indexName,
	}, nil
}

func (i *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	var req IndexRequest

	bodyBuf := new(bytes.Buffer)
	_, _ = bodyBuf.ReadFrom(r.Body)

	if bodyBuf.Len() == 0 {
		i.logger.Println("request body is empty")
		response.WriteJSONError(w, "request", "Request body is empty", http.StatusBadRequest)
		return
	}

	err := json.Unmarshal(bodyBuf.Bytes(), &req)
	if err != nil {
		i.logger.Println(err.Error())
		response.WriteJSONError(w, "request", "Unable to unmarshal JSON request", http.StatusBadRequest)
		return
	}

	validationErrs := req.Validate()
	if len(validationErrs) > 0 {
		i.logger.Println("Request failed validation", validationErrs)
		response.WriteJSONErrors(w, "Some fields have failed validation", validationErrs, http.StatusBadRequest)
		return
	}

	response := &indexResponse{}

	if err := i.doIndex(i.indexName, response, req.Persons); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	jsonResp, _ := json.Marshal(response)

	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write(jsonResp)

	i.logger.Println("Request took: ", time.Since(start))
}

func (i *IndexHandler) doIndex(indexName string, response *indexResponse, persons []Person) error {
	op := elasticsearch.NewBulkOp(indexName)

	for _, p := range persons {
		err := op.Index(p.Id(), p)

		if err == elasticsearch.ErrOpTooLarge {
			response.Add(i.client.DoBulk(op))
			op.Reset()
			err = op.Index(p.Id(), p)
		}

		if err != nil {
			i.logger.Println(err)

			return fmt.Errorf("could not construct index request for id=%d", p.Id())
		}
	}

	if !op.Empty() {
		response.Add(i.client.DoBulk(op))
	}

	return nil
}

type indexResponse struct {
	Successful int      `json:"successful"`
	Failed     int      `json:"failed"`
	Errors     []string `json:"errors,omitempty"`
}

func (r *indexResponse) Add(result elasticsearch.BulkResult, err error) {
	r.Successful += result.Successful
	r.Failed += result.Failed

	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}
}
