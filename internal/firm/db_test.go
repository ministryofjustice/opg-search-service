package firm

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
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	_, err = conn.Exec(ctx, `
		INSERT INTO firm (id, firmname, addressline1, town, county, postcode, phonenumber, email, firmnumber)
		VALUES (1, 'firm co', '123 Fake street', 'Faketon', 'Fakeshire', 'F1 1FF', '0111111111', 'email@example.com', '1234'),
		(2, 'firm co', '123 Fake street', 'Faketon', 'Fakeshire', 'F1 1FF', '0111111111', 'email@example.com', '1234'),
		(3, 'firm co', '123 Fake street', 'Faketon', 'Fakeshire', 'F1 1FF', '0111111111', 'email@example.com', '1234');
	`)
	if !assert.Nil(err) {
		return
	}

	db := DB{conn: conn}

	min, max, err := db.QueryIDRange(ctx)
	assert.Nil(err)
	assert.Equal(1, min)
	assert.Equal(3, max)
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
	if !assert.Nil(err) {
		return
	}

	_, err = conn.Exec(ctx, `
		INSERT INTO firm (id, firmname, addressline1, town, county, postcode, phonenumber, email, firmnumber)
		VALUES (1, 'firm co', '123 Fake street', 'Faketon', 'Fakeshire', 'F1 1FF', '0111111111', 'email@example.com', 1234),
		(2, 'firm co', '123 Fake street', 'Faketon', 'Fakeshire', 'F1 1FF', '0111111111', 'email@example.com', 1234),
		(3, '', '', '', '', '', '', '', 1),
		(4, 'firm co', '123 Fake street', 'Faketon', 'Fakeshire', 'F1 1FF', '0111111111', 'email@example.com', 1234);
	`)
	if !assert.Nil(err) {
		return
	}

	resultsCh := make(chan index.Indexable)
	db := DB{conn: conn}

	go func() {
		err = db.QueryByID(ctx, resultsCh, 1, 3)
		assert.Nil(err)
	}()

	first, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Firm{
		ID:           i64(1),
		Persontype:   "Firm",
		Email:        "email@example.com",
		FirmName:     "firm co",
		FirmNumber:   "1234",
		AddressLine1: "123 Fake street",
		Town:         "Faketon",
		County:       "Fakeshire",
		Postcode:     "F1 1FF",
		Phonenumber:  "0111111111",
	}, first)

	second, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Firm{
		ID:           i64(2),
		Persontype:   "Firm",
		Email:        "email@example.com",
		FirmName:     "firm co",
		FirmNumber:   "1234",
		AddressLine1: "123 Fake street",
		Town:         "Faketon",
		County:       "Fakeshire",
		Postcode:     "F1 1FF",
		Phonenumber:  "0111111111",
	}, second)

	third, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Firm{
		ID:         i64(3),
		Persontype: "Firm",
		FirmNumber: "1",
	}, third)

	_, ok = read(resultsCh, time.Nanosecond)
	assert.False(ok)
}

func i64(x int) *int64 {
	y := int64(x)
	return &y
}
