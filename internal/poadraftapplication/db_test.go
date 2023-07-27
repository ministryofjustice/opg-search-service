package poadraftapplication

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/index"
	"github.com/stretchr/testify/assert"
)

const connectionString = "postgres://searchservice:searchservice@postgres:5432/searchservice"

func read(results chan index.Indexable, deadline time.Duration) (index.Indexable, bool) {
	select {
	case x := <-results:
		return x, true
	case <-time.After(deadline):
		return nil, false
	}
}

func TestGetIDRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgres test")
		return
	}

	assert := assert.New(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connectionString)
	assert.Nil(err)

	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	assert.Nil(err)

	_, err = conn.Exec(ctx, `
		INSERT INTO poa.draft_applications (id, donorname)
		VALUES (1, 'TLane Araxa'),
		(2, 'MMosa SepIch');
	`)
	assert.Nil(err)

	db := DB{conn: conn}

	min, max, err := db.QueryIDRange(ctx)
	assert.Nil(err)
	assert.Equal(1, min)
	assert.Equal(2, max)
}

func TestQueryByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgres test")
		return
	}

	assert := assert.New(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connectionString)
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	assert.Nil(err)

	_, err = conn.Exec(ctx, `
		INSERT INTO poa.draft_applications (id, donorname)
		VALUES (1, 'TLane Araxa'),
		(2, 'MMosa SepIch');
	`)
	assert.Nil(err)

	resultsCh := make(chan index.Indexable)
	db := DB{conn: conn}

	go func() {
		err = db.QueryByID(ctx, resultsCh, 1, 3)
		assert.Nil(err)
	}()

	first, ok := read(resultsCh, time.Second)
	assert.True(ok)

	assert.Equal(DraftApplication{
		ID: i64(1),
		DonorName: "TLane Araxa",
	}, first)

	second, ok := read(resultsCh, time.Second)
	assert.True(ok)

	assert.Equal(DraftApplication{
		ID: i64(2),
		DonorName: "MMosa SepIch",
	}, second)

	_, ok = read(resultsCh, time.Nanosecond)
	assert.False(ok)
}

func i64(x int) *int64 {
	y := int64(x)
	return &y
}
