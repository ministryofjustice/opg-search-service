package search

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type SearchClient interface {
	Search(ctx context.Context, indices []string, requestBody map[string]interface{}) (*elasticsearch.SearchResult, error)
}

type PrepareQuery func(*Request) map[string]interface{}

type Handler struct {
	logger       *logrus.Logger
	client       SearchClient
	indices      []string
	prepareQuery PrepareQuery
}

func NewHandler(logger *logrus.Logger, client SearchClient, indices []string, prepareQuery PrepareQuery) *Handler {
	return &Handler{
		logger:       logger,
		client:       client,
		indices:      indices,
		prepareQuery: prepareQuery,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	req, err := parseSearchRequest(r)
	if err != nil {
		h.logger.Println(err)
		response.WriteJSONError(w, "request", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.client.Search(r.Context(), h.indices, h.prepareQuery(req))
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, context.Canceled) {
			code = 499
		}
		h.logger.Println(err.Error())
		response.WriteJSONError(w, "request", "unexpected error from elasticsearch", code)
		return
	}

	resp := Response{
		Aggregations: result.Aggregations,
		Results:      result.Hits,
		Total: ResponseTotal{
			Count: result.Total,
			Exact: result.TotalExact,
		},
	}

	jsonResp, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)

	h.logger.Printf("Request took: %d", time.Since(start))
}
