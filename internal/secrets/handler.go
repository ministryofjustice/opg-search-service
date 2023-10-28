package secrets

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type SecretsCache interface {
	ClearCache()
}

type Handler struct {
	logger *logrus.Logger
	cache  SecretsCache
}

func NewHandler(logger *logrus.Logger, cache SecretsCache) *Handler {
	return &Handler{
		logger: logger,
		cache:  cache,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.cache.ClearCache()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}
