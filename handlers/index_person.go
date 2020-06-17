package handlers

import (
	"log"
	"net/http"
)

type IndexPersonHandler struct {
	logger *log.Logger
}

func NewIndexPersonHandler(logger *log.Logger) (*IndexPersonHandler, error) {
	return &IndexPersonHandler{
		logger: logger,
	}, nil
}

func (i IndexPersonHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	i.logger.Println("indexing...")
}
