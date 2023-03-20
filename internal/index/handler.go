package index

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type IndexClient interface {
	DoBulk(ctx context.Context, op *elasticsearch.BulkOp) (elasticsearch.BulkResult, error)
}

type Parser func([]byte) (Validatable, error)

type Handler struct {
	logger  *logrus.Logger
	client  IndexClient
	indices []string
	parser  Parser
}

func NewHandler(logger *logrus.Logger, client IndexClient, indices []string, parser Parser) *Handler {
	return &Handler{
		logger:  logger,
		client:  client,
		indices: indices,
		parser:  parser,
	}
}

func (i *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	bodyBuf := new(bytes.Buffer)
	_, _ = bodyBuf.ReadFrom(r.Body)

	if bodyBuf.Len() == 0 {
		i.logger.Println("request body is empty")
		response.WriteJSONError(w, "request", "Request body is empty", http.StatusBadRequest)
		return
	}

	req, err := i.parser(bodyBuf.Bytes())
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

	for _, index := range i.indices {
		if err := i.doIndex(r.Context(), index, response, req.Items()); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}

	jsonResp, _ := json.Marshal(response)

	w.WriteHeader(http.StatusAccepted)

	_, _ = w.Write(jsonResp)

	i.logger.Println("Request took: ", time.Since(start))
}

func (i *Handler) doIndex(ctx context.Context, indexName string, response *indexResponse, items []Indexable) error {
	op := elasticsearch.NewBulkOp(indexName)

	for _, f := range items {
		err := op.Index(f.Id(), f)

		if err == elasticsearch.ErrOpTooLarge {
			response.Add(i.client.DoBulk(ctx, op))
			op.Reset()
			err = op.Index(f.Id(), f)
		}

		if err != nil {
			i.logger.Println(err)

			return fmt.Errorf("could not construct index request for id=%d", f.Id())
		}
	}

	if !op.Empty() {
		response.Add(i.client.DoBulk(ctx, op))
	}

	return nil
}
