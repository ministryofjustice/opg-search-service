package cli

import (
	"flag"
	"log"
	"net/http"
	"os"
)

type healthCheck struct {
	logger    *log.Logger
	shouldRun *bool
	checkUrl  string
	exit      func(code int)
}

func NewHealthCheck(logger *log.Logger) *healthCheck {
	return &healthCheck{
		logger:   logger,
		checkUrl: "http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check",
		exit:     os.Exit,
	}
}

func (h *healthCheck) DefineFlags() {
	h.shouldRun = flag.Bool("hc", false, "perform a health check")
}

func (h *healthCheck) ShouldRun() bool {
	return *h.shouldRun
}

func (h *healthCheck) Run() {
	resp, err := http.Get(h.checkUrl)
	if err != nil || resp.StatusCode != 200 {
		h.logger.Println("FAIL")
		h.exit(1)
		return
	}
	h.logger.Println("OK")
	h.exit(0)
}
