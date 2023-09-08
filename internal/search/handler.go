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

type PrepareQuery func(*Request) ([]string, map[string]interface{})

type Handler struct {
	logger       *logrus.Logger
	client       SearchClient
	prepareQuery PrepareQuery
}

func NewHandler(logger *logrus.Logger, client SearchClient, prepareQuery PrepareQuery) *Handler {
	return &Handler{
		logger:       logger,
		client:       client,
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

	indices, requestBody := h.prepareQuery(req)

	result, err := h.client.Search(r.Context(), indices, requestBody)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			response.WriteJSONError(w, "request", "search request was cancelled", 499)
		} else {
			response.WriteJSONError(w, "request", "unexpected error from elasticsearch", http.StatusInternalServerError)
		}
		h.logger.Println(err.Error())

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
