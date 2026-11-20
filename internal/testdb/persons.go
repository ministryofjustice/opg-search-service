package testdb

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

// CreatePersonTables uses session-local copies of the shared person fixtures.
func CreatePersonTables(ctx context.Context, conn *pgx.Conn) error {
	schema, err := os.ReadFile("../testdata/schema.sql")
	if err != nil {
		return err
	}
	personSchema, _, ok := strings.Cut(string(schema), "-- Table: supervision.firm")
	if !ok {
		return errors.New("missing firm section in the shared test schema")
	}
	personSchema = strings.NewReplacer(
		"CREATE TABLE public.", "CREATE TEMP TABLE pg_temp.",
		"DROP TABLE IF EXISTS public.", "DROP TABLE IF EXISTS pg_temp.",
	).Replace(personSchema)
	_, err = conn.Exec(ctx, personSchema)
	return err
}
