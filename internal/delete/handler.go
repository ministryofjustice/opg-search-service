package delete

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type DeleteClient interface {
	Delete(indices []string, requestBody map[string]interface{}) (*elasticsearch.DeleteResult, error)
}

type PrepareQuery func(*Request) map[string]interface{}

type Handler struct {
	logger  *logrus.Logger
	client  DeleteClient
	indices []string
}

func NewHandler(logger *logrus.Logger, client DeleteClient, indices []string) *Handler {
	return &Handler{
		logger:  logger,
		client:  client,
		indices: indices,
	}
}

func prepareQuery(req *Request) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"uId": req.Uid,
			},
		},
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	req, err := parseDeleteRequest(r)
	if err != nil {
		h.logger.Println(err)
		response.WriteJSONError(w, "request", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.client.Delete(h.indices, prepareQuery(req))
	if err != nil {
		h.logger.Println(err.Error())
		response.WriteJSONError(w, "request", "unexpected error from elasticsearch", http.StatusInternalServerError)
		return
	}

	resp := Response{
		Total: result.Total,
	}

	jsonResp, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jsonResp)

	h.logger.Printf("Request took: %d", time.Since(start))
}
