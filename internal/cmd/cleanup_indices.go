package cmd

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus"
)

type CleanupIndicesClient interface {
	ResolveAlias(ctx context.Context, alias string) (string, error)
	Indices(ctx context.Context, term string) ([]string, error)
	DeleteIndex(ctx context.Context, name string) error
}

type cleanupIndicesCommand struct {
	logger         *logrus.Logger
	client         CleanupIndicesClient
	currentIndices map[string][]byte
}

func NewCleanupIndices(logger *logrus.Logger, client CleanupIndicesClient, indices map[string][]byte) *cleanupIndicesCommand {
	return &cleanupIndicesCommand{
		logger:         logger,
		client:         client,
		currentIndices: indices,
	}
}

func (c *cleanupIndicesCommand) Info() (name, description string) {
	return "cleanup-indices", "remove unused indices"
}

func (c *cleanupIndicesCommand) Run(args []string) error {
	ctx := context.Background()
	flagset := flag.NewFlagSet("cleanup-indices", flag.ExitOnError)

	explain := flagset.Bool("explain", false, "explain the changes that will be made")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	for _, aliasName := range []string{firm.AliasName, person.AliasName} {
		aliasedIndex, err := c.client.ResolveAlias(ctx, aliasName)
		if err != nil {
			return err
		}
		if _, ok := c.currentIndices[aliasedIndex]; !ok {
			return fmt.Errorf("alias '%s' is set to '%s' not a current index: %s", aliasName, aliasedIndex, mapKeys(c.currentIndices))
		}

		indices, err := c.client.Indices(ctx, aliasName+"_*")
		if err != nil {
			return err
		}

		for _, index := range indices {
			if _, ok := c.currentIndices[index]; !ok {
				if *explain {
					c.logger.Println("will delete", index)
				} else {
					if err := c.client.DeleteIndex(ctx, index); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func mapKeys(m map[string][]byte) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
