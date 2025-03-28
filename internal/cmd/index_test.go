package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/elasticsearch"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestIndexPerson(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping indexing test")
		return
	}

	assert := assert.New(t)
	ctx := context.Background()

	mockClient := &elasticsearch.MockESClient{}

	mockClient.
		On("DoBulk", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(elasticsearch.BulkResult{
			Successful: 0,
			Failed:     0,
			Error:      "",
		}, nil)

	l, hook := test.NewNullLogger()
	command := NewIndex(l, mockClient, nil, []IndexConfig{
		{
			Name:   "person_1",
			Alias:  "person",
			Config: indexConfig,
		},
	})

	_ = os.Setenv("SEARCH_SERVICE_DB_PASS", "searchservice")
	_ = os.Setenv("SEARCH_SERVICE_DB_USER", "searchservice")
	_ = os.Setenv("SEARCH_SERVICE_DB_HOST", "postgres")
	_ = os.Setenv("SEARCH_SERVICE_DB_PORT", "5432")
	_ = os.Setenv("SEARCH_SERVICE_DB_DATABASE", "searchservice")

	connString, _ := command.dbConnectionString()

	conn, err := pgx.Connect(ctx, connString)
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx) //nolint:errcheck // no need to check DB close error in tests

	schemaSql, _ := os.ReadFile("../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	err = command.Run([]string{"--person"})
	assert.Nil(err)
	assert.Equal("indexing done successful=0 failed=0", hook.LastEntry().Message)
}

func TestIndexFirm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping indexing test")
		return
	}

	assert := assert.New(t)
	ctx := context.Background()

	l, hook := test.NewNullLogger()
	command := NewIndex(l, &elasticsearch.MockESClient{}, nil, []IndexConfig{
		{
			Name:   "firm_1",
			Alias:  "firm",
			Config: indexConfig,
		},
	})

	_ = os.Setenv("SEARCH_SERVICE_DB_PASS", "searchservice")
	_ = os.Setenv("SEARCH_SERVICE_DB_USER", "searchservice")
	_ = os.Setenv("SEARCH_SERVICE_DB_HOST", "postgres")
	_ = os.Setenv("SEARCH_SERVICE_DB_PORT", "5432")
	_ = os.Setenv("SEARCH_SERVICE_DB_DATABASE", "searchservice")

	connString, _ := command.dbConnectionString()

	conn, err := pgx.Connect(ctx, connString)
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx) //nolint:errcheck // no need to check DB close error in tests

	schemaSql, _ := os.ReadFile("../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	err = command.Run([]string{"--firm"})
	assert.Nil(err)
	assert.Equal("indexing done successful=0 failed=0", hook.LastEntry().Message)
}

func TestNewIndexConfig(t *testing.T) {
	l, _ := test.NewNullLogger()
	ic := NewIndexConfig(func() ([]byte, error) { return []byte{}, nil }, "somealias", l)
	assert.Regexp(t, `[a-z]+_[a-z0-9]+`, ic.Name)
}
