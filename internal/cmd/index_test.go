package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

// TODO flaky tests when ran locally
func TestIndexPerson(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	l, hook := test.NewNullLogger()
	command := NewIndex(l, &elasticsearch.MockESClient{}, nil, map[string][]byte{"person_1": indexConfig})

	os.Setenv("SEARCH_SERVICE_DB_PASS", "searchservice")
	os.Setenv("SEARCH_SERVICE_DB_USER", "searchservice")
	os.Setenv("SEARCH_SERVICE_DB_HOST", "postgres")
	os.Setenv("SEARCH_SERVICE_DB_PORT", "5432")
	os.Setenv("SEARCH_SERVICE_DB_DATABASE", "searchservice")

	connString, _ := command.dbConnectionString()

	conn, err := pgx.Connect(ctx, connString)
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("./testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	err = command.Run([]string{})
	assert.Nil(err)
	assert.Equal("indexing done successful=0 failed=0", hook.LastEntry().Message)
}

func TestIndexFirm(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	l, hook := test.NewNullLogger()
	command := NewIndex(l, &elasticsearch.MockESClient{}, nil, map[string][]byte{"test-index": indexConfig})

	os.Setenv("SEARCH_SERVICE_DB_PASS", "searchservice")
	os.Setenv("SEARCH_SERVICE_DB_USER", "searchservice")
	os.Setenv("SEARCH_SERVICE_DB_HOST", "postgres")
	os.Setenv("SEARCH_SERVICE_DB_PORT", "5432")
	os.Setenv("SEARCH_SERVICE_DB_DATABASE", "searchservice")

	connString, _ := command.dbConnectionString()

	conn, err := pgx.Connect(ctx, connString)
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("./testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	err = command.Run([]string{"--firm"})
	assert.Nil(err)
	assert.Equal("indexing done successful=0 failed=0", hook.LastEntry().Message)
}
