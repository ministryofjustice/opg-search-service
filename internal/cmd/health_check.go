package cmd

import (
	"errors"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

type healthCheckCommand struct {
	logger   *logrus.Logger
	checkUrl string
}

func NewHealthCheck(logger *logrus.Logger) *healthCheckCommand {
	return &healthCheckCommand{
		logger:   logger,
		checkUrl: "http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check",
	}
}

func (h *healthCheckCommand) Name() string {
	return "hc"
}

func (h *healthCheckCommand) Run(args []string) error {
	resp, err := http.Get(h.checkUrl)
	if err != nil || resp.StatusCode != 200 {
		return errors.New("FAIL")
	}
	h.logger.Println("OK")
	return nil
}
