package delete

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/ministryofjustice/opg-search-service/internal/response"
	"github.com/sirupsen/logrus"
)

type DeleteClient interface {
	Delete(indices []string, requestBody map[string]interface{}) (*elasticsearch.DeleteResult, error)
}

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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	if uid == "" {
		err := errors.New("uid is required and cannot be empty")
		h.logger.Println(err)
		response.WriteJSONError(w, "request", err.Error(), http.StatusBadRequest)
		return
	}

	requestBody := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"uId": uid,
			},
		},
	}

	result, err := h.client.Delete(h.indices, requestBody)
	if err != nil {
		h.logger.Println(err.Error())
		response.WriteJSONError(w, "request", "unexpected error from elasticsearch", http.StatusInternalServerError)
		return
	}

	if result.Total == 1 {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	} else if result.Total == 0 {
		response.WriteJSONError(w, "request", "could not find document to delete", http.StatusNotFound)
	} else {
		response.WriteJSONError(w, "request", fmt.Sprintf("deleted %d documents matching query", result.Total), http.StatusInternalServerError)
	}
}
