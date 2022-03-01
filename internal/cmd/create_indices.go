package cmd

import (
	"github.com/sirupsen/logrus"
)

type createIndicesCommand struct {
	logger *logrus.Logger
}

func NewCreateIndices(logger *logrus.Logger) *createIndicesCommand {
	return &createIndicesCommand{
		logger: logger,
	}
}

func (c *createIndicesCommand) Name() string {
	return "create-indices"
}

func (c *createIndicesCommand) Run(args []string) error {
	c.logger.Print("create-indices is now a no-op, the only thing it does is log this message")

	return nil
}
