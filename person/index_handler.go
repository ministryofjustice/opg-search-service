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

const indexBatchSize = 40000

type IndexHandler struct {
	logger *logrus.Logger
	es     elasticsearch.ClientInterface
}

func NewIndexHandler(logger *logrus.Logger) (*IndexHandler, error) {
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

func (i IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	op := elasticsearch.NewBulkOp(personIndexName)
	var results []elasticsearch.IndexResult

	for _, p := range req.Persons {
		err := op.Index(p.Id(), p)

		if err == elasticsearch.ErrOpTooLarge {
			results = append(results, i.es.DoBulk(op)...)
			op.Reset()
			err = op.Index(p.Id(), p)

		} else if err != nil {
			i.logger.Println(err)

			http.Error(w, fmt.Sprintf("could not construct index request for id=%d", p.Id()), http.StatusBadRequest)
			return
		}
	}

	if !op.Empty() {
		results = append(results, i.es.DoBulk(op)...)
	}

	jsonResp, _ := json.Marshal(response.IndexResponse{Results: results})

	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write(jsonResp)

	i.logger.Println("Request took: ", time.Since(start))
}
