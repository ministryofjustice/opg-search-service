package deputy

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
		INSERT INTO persons (id, uid, caserecnumber, email, dob, firstname, middlenames, surname, companyname, type, organisationname)
		VALUES (1, 700656728331, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'actor_deputy', 'Orgz'),
		(2, 700656728332, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'lpa_donor', null),
		(3, 700656728333, null, null, '1990-01-02', 'J', null, 'J', null, 'lpa_donor', null);
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
		INSERT INTO persons (id, uid, caserecnumber, deputynumber, email, dob, firstname, middlenames, surname, companyname, type, organisationname)
		VALUES (1, 700656728331, 1010101, null, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'actor_deputy', 'Orgz'),
		(2, 700656728332, null, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'actor_deputy', null),
		(3, 700656728333, null, 12345, 'deputy@example.com', '1970-03-06', 'D', null, 'D', null, 'actor_deputy', null),
		(4, 700656728334, null, null, null, '1990-01-02', 'J', null, 'J', null, 'actor_deputy', null);
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
	assert.Equal(Deputy{
		ID:               i64(1),
		UID:              "7006-5672-8331",
		Normalizeduid:    700656728331,
		DeputyNumber:     nil,
		Dob:              "02/01/2002",
		Firstname:        "John",
		Middlenames:      "J",
		Surname:          "Johnson",
		CompanyName:      "& co",
		Persontype:       "Deputy",
		OrganisationName: "Orgz",
	}, first)

	second, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Deputy{
		ID:               i64(2),
		UID:              "7006-5672-8332",
		Normalizeduid:    700656728332,
		DeputyNumber:     nil,
		Dob:              "02/01/1990",
		Firstname:        "Jack",
		Middlenames:      "",
		Surname:          "Jackson",
		CompanyName:      "",
		Persontype:       "Deputy",
		OrganisationName: "",
	}, second)

	third, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Deputy{
		ID:               i64(3),
		UID:              "7006-5672-8333",
		Normalizeduid:    700656728333,
		DeputyNumber:     i64(12345),
		Dob:              "06/03/1970",
		Firstname:        "D",
		Middlenames:      "",
		Surname:          "D",
		CompanyName:      "",
		Persontype:       "Deputy",
		OrganisationName: "",
	}, third)

	_, ok = read(resultsCh, time.Nanosecond)
	assert.False(ok)
}

func TestQueryFromDate(t *testing.T) {
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
		INSERT INTO persons (id, updateddate, uid, caserecnumber, email, dob, firstname, middlenames, surname, companyname, type, organisationname)
		VALUES (1, '2021-01-03 12:00:00', 7000, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'actor_deputy', 'Orgz'),
		(2, '2021-01-02 12:00:00', 7002, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'actor_deputy', null),
		(3, '2021-01-01 12:00:00', 7003, null, null, '1990-01-02', 'J', null, 'J', null, 'actor_deputy', null);
	`)
	if !assert.Nil(err) {
		return
	}

	resultsCh := make(chan index.Indexable)
	db := DB{conn: conn}

	go func() {
		err = db.QueryFromDate(ctx, resultsCh, time.Date(2021, time.January, 1, 23, 0, 0, 0, time.UTC))
		assert.Nil(err)
	}()

	first, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Deputy{
		ID:               i64(1),
		UID:              "7000",
		Normalizeduid:    7000,
		DeputyNumber:     nil,
		Dob:              "02/01/2002",
		Firstname:        "John",
		Middlenames:      "J",
		Surname:          "Johnson",
		CompanyName:      "& co",
		Persontype:       "Deputy",
		OrganisationName: "Orgz",
	}, first)

	second, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Deputy{
		ID:               i64(2),
		UID:              "7002",
		Normalizeduid:    7002,
		DeputyNumber:     nil,
		Dob:              "02/01/1990",
		Firstname:        "Jack",
		Middlenames:      "",
		Surname:          "Jackson",
		CompanyName:      "",
		Persontype:       "Deputy",
		OrganisationName: "",
	}, second)

	_, ok = read(resultsCh, time.Nanosecond)
	assert.False(ok)
}
