package cmd

import (
	"errors"
	"flag"
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

func (h *healthCheckCommand) Info() (name, description string) {
	return "hc", "run healthcheck"
}

func (h *healthCheckCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("hc", flag.ExitOnError)

	if err := flagset.Parse(args); err != nil {
		return err
	}

	resp, err := http.Get(h.checkUrl)
	if err != nil || resp.StatusCode != 200 {
		return errors.New("FAIL")
	}
	h.logger.Println("OK")
	return nil
}
