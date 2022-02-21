package firm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type IndexClient interface {
	DoBulk(op *elasticsearch.BulkOp) (elasticsearch.BulkResult, error)
}

type IndexHandler struct {
	logger  *logrus.Logger
	client  IndexClient
	indices []string
}

func NewIndexHandler(logger *logrus.Logger, client IndexClient, indices []string) *IndexHandler {
	return &IndexHandler{
		logger:  logger,
		client:  client,
		indices: indices,
	}
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
	i.logger.Println("&req")
	i.logger.Println(&req)
	i.logger.Println("r.Body")
	i.logger.Println(r.Body)
	i.logger.Println("bodyBuf.Bytes()")
	i.logger.Println(bodyBuf.Bytes())
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

	i.logger.Println("Response")
	i.logger.Println(response)

	for _, index := range i.indices {
		i.logger.Println("index in for loop")
		i.logger.Print(index)
		i.logger.Println("req.Firms")
		i.logger.Println(req.Firms)
		if err := i.doIndex(index, response, req.Firms); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	jsonResp, _ := json.Marshal(response)
	i.logger.Println("jsonResp")
	i.logger.Println(jsonResp)


	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write(jsonResp)

	i.logger.Println("Request took: ", time.Since(start))
}

func (i *IndexHandler) doIndex(indexName string, response *indexResponse, firms []Firm) error {
	i.logger.Println("In the doIndex")
	op := elasticsearch.NewBulkOp(indexName)
	i.logger.Println("op")
	i.logger.Println(op)
	for _, f := range firms {
		err := op.Index(f.Id(), f)

		if err == elasticsearch.ErrOpTooLarge {
			response.Add(i.client.DoBulk(op))
			op.Reset()
			err = op.Index(f.Id(), f)
		}

		if err != nil {
			i.logger.Println(err)

			return fmt.Errorf("could not construct index request for id=%d", f.Id())
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
