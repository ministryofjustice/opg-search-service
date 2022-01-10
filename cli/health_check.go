package cli

import (
	"errors"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

type healthCheck struct {
	logger   *logrus.Logger
	checkUrl string
}

func NewHealthCheck(logger *logrus.Logger) *healthCheck {
	return &healthCheck{
		logger:   logger,
		checkUrl: "http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check",
	}
}

func (h *healthCheck) Name() string {
	return "hc"
}

func (h *healthCheck) Run(args []string) error {
	resp, err := http.Get(h.checkUrl)
	if err != nil || resp.StatusCode != 200 {
		return errors.New("FAIL")
	}
	h.logger.Println("OK")
	return nil
}
