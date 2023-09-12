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
		INSERT INTO cases (id, uid, caserecnumber)
		VALUES (101, 101, 'M-7QQQ-P4DF-4UX3'),
		(102, 102, 'M-8YYY-P4DF-4UX3');
	`)
	assert.Nil(err)

	_, err = conn.Exec(ctx, `
		INSERT INTO poa.draft_applications (id, lpa_id, donorfirstnames, donorlastname, donordob, donorpostcode)
		VALUES (1, 101, 'TLane', 'Araxa', '12/12/2000', 'X11 11X'),
		(2, 102, 'MMosa', 'SepIch', '09/09/1999', 'Y11 11Y');
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
		INSERT INTO cases (id, uid, caserecnumber)
		VALUES (101, 101, 'M-7QQQ-P4DF-4UX3'),
		(102, 102, 'M-8YYY-P4DF-4UX3');
	`)
	assert.Nil(err)

	_, err = conn.Exec(ctx, `
		INSERT INTO poa.draft_applications (id, lpa_id, donorfirstnames, donorlastname, donordob, donorpostcode)
		VALUES (1, 101, 'TLane', 'Araxa', '12/12/2000', 'X11 11X'),
		(2, 102, 'MMosa', 'SepIch', '09/09/1999', 'Y11 11Y');
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

	firstUid := "M-7QQQ-P4DF-4UX3"
	assert.Equal(DraftApplication{
		UID: firstUid,
		Donor: DraftApplicationDonor{
			FirstNames: "TLane",
			LastName:   "Araxa",
			Dob:        "12/12/2000",
			Postcode:   "X11 11X",
		},
	}, first)

	second, ok := read(resultsCh, time.Second)
	assert.True(ok)

	secondUid := "M-8YYY-P4DF-4UX3"
	assert.Equal(DraftApplication{
		UID: secondUid,
		Donor: DraftApplicationDonor{
			FirstNames: "MMosa",
			LastName:   "SepIch",
			Dob:        "09/09/1999",
			Postcode:   "Y11 11Y",
		},
	}, second)

	_, ok = read(resultsCh, time.Nanosecond)
	assert.False(ok)
}
