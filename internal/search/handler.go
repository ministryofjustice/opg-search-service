package search

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type SearchClient interface {
	Search(indices []string, requestBody map[string]interface{}) (*elasticsearch.SearchResult, error)
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
	req, err := parseSearchRequest(r)
	if err != nil {
		h.logger.Println(err)
		response.WriteJSONError(w, "request", err.Error(), http.StatusBadRequest)
		return
	}

	_, subsegment := xray.BeginSubsegment(r.Context(), "do-search")
	result, err := h.client.Search(h.indices, h.prepareQuery(req))
	subsegment.Close(nil)

	if err != nil {
		h.logger.Println(err.Error())
		response.WriteJSONError(w, "request", "unexpected error from elasticsearch", http.StatusInternalServerError)
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
}
