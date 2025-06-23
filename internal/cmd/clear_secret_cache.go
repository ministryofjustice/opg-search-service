package cmd

import (
	"errors"
	"flag"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type clearSecretCacheCommand struct {
	logger *logrus.Logger
	url    string
}

func NewClearSecretCache(logger *logrus.Logger) *clearSecretCacheCommand {
	return &clearSecretCacheCommand{
		logger: logger,
		url:    "http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/clear-secret-cache",
	}
}

func (c *clearSecretCacheCommand) Info() (name, description string) {
	return "clear-secret-cache", "clear secret cache"
}

func (c *clearSecretCacheCommand) Run(args []string) error {
	flagset := flag.NewFlagSet("clear-secret-cache", flag.ExitOnError)

	if err := flagset.Parse(args); err != nil {
		return err
	}

	resp, err := http.Get(c.url)
	if err != nil || resp.StatusCode != 200 {
		return errors.New("FAIL")
	}
	c.logger.Println("Secrets cache cleared")
	return nil
}
