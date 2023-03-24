package person

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
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
		VALUES (1, 700656728331, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'lpa_donor', 'Orgz'),
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
		VALUES (1, 700656728331, 1010101, null, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'lpa_donor', 'Orgz'),
		(2, 700656728332, null, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'lpa_donor', null),
		(3, 700656728333, null, 12345, 'deputy@example.com', '1970-03-06', 'D', null, 'D', null, 'actor_deputy', null),
		(4, 700656728334, null, null, null, '1990-01-02', 'J', null, 'J', null, 'lpa_donor', null);

		INSERT INTO phonenumbers (id, person_id, phone_number)
		VALUES (1, 1, '077777777');

		INSERT INTO addresses (id, person_id, address_lines, postcode)
		VALUES (1, 1, json_build_object('0', '1 Road', '2', 'Place'), 'S1 1AB');

		INSERT INTO cases (id, uid, caserecnumber, onlinelpaid, batchid, casetype, casesubtype)
		VALUES (1, 700656728311, '545534', 'A123', 'x', 'lpa', 'hw'),
		(2, 700656728312, '545532', 'A124', 'y', 'lpa', 'pfa');

		INSERT INTO person_caseitem (person_id, caseitem_id) VALUES (1, 1), (1, 2);
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
	assert.Equal(Person{
		ID:               i64(1),
		UID:              "7006-5672-8331",
		Normalizeduid:    700656728331,
		CaseRecNumber:    "1010101",
		DeputyNumber:     nil,
		Email:            "email@example.com",
		Dob:              "02/01/2002",
		Firstname:        "John",
		Middlenames:      "J",
		Surname:          "Johnson",
		CompanyName:      "& co",
		Persontype:       "Donor",
		OrganisationName: "Orgz",
		Phonenumbers: []PersonPhonenumber{{
			Phonenumber: "077777777",
		}},
		Addresses: []PersonAddress{{
			Addresslines: []string{"1 Road", "", "Place"},
			Postcode:     "S1 1AB",
		}},
		Cases: []PersonCase{{
			UID:           "7006-5672-8311",
			Normalizeduid: 700656728311,
			Caserecnumber: "545534",
			OnlineLpaId:   "A123",
			Batchid:       "x",
			Casetype:      "lpa",
			Casesubtype:   "hw",
		}, {
			UID:           "7006-5672-8312",
			Normalizeduid: 700656728312,
			Caserecnumber: "545532",
			OnlineLpaId:   "A124",
			Batchid:       "y",
			Casetype:      "lpa",
			Casesubtype:   "pfa",
		}},
	}, first)

	second, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Person{
		ID:               i64(2),
		UID:              "7006-5672-8332",
		Normalizeduid:    700656728332,
		CaseRecNumber:    "",
		DeputyNumber:     nil,
		Email:            "",
		Dob:              "02/01/1990",
		Firstname:        "Jack",
		Middlenames:      "",
		Surname:          "Jackson",
		CompanyName:      "",
		Persontype:       "Donor",
		OrganisationName: "",
		Phonenumbers:     nil,
		Addresses:        nil,
		Cases:            nil,
	}, second)

	third, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Person{
		ID:               i64(3),
		UID:              "7006-5672-8333",
		Normalizeduid:    700656728333,
		CaseRecNumber:    "",
		DeputyNumber:     i64(12345),
		Email:            "deputy@example.com",
		Dob:              "06/03/1970",
		Firstname:        "D",
		Middlenames:      "",
		Surname:          "D",
		CompanyName:      "",
		Persontype:       "Deputy",
		OrganisationName: "",
		Phonenumbers:     nil,
		Addresses:        nil,
		Cases:            nil,
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
		VALUES (1, '2021-01-03 12:00:00', 7000, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'lpa_attorney', 'Orgz'),
		(2, '2021-01-02 12:00:00', 7002, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'lpa_donor', null),
		(3, '2021-01-01 12:00:00', 7003, null, null, '1990-01-02', 'J', null, 'J', null, 'lpa_donor', null);

		INSERT INTO phonenumbers (id, person_id, phone_number)
		VALUES (1, 1, '077777777');

		INSERT INTO addresses (id, person_id, address_lines, postcode)
		VALUES (1, 1, json_build_array('123 Fake St'), 'S1 1AB');

		INSERT INTO cases (id, uid, caserecnumber, onlinelpaid, batchid, casetype, casesubtype)
		VALUES (1, 7000, '545534', 'A123', 'x', 'lpa', 'hw'),
		(2, 7002, '545532', 'A124', 'y', 'lpa', 'pfa');

		INSERT INTO person_caseitem (person_id, caseitem_id) VALUES (1, 1), (1, 2);
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
	assert.Equal(Person{
		ID:               i64(1),
		UID:              "7000",
		Normalizeduid:    7000,
		CaseRecNumber:    "1010101",
		DeputyNumber:     nil,
		Email:            "email@example.com",
		Dob:              "02/01/2002",
		Firstname:        "John",
		Middlenames:      "J",
		Surname:          "Johnson",
		CompanyName:      "& co",
		Persontype:       "Attorney",
		OrganisationName: "Orgz",
		Phonenumbers: []PersonPhonenumber{{
			Phonenumber: "077777777",
		}},
		Addresses: []PersonAddress{{
			Addresslines: []string{"123 Fake St"},
			Postcode:     "S1 1AB",
		}},
		Cases: []PersonCase{{
			UID:           "7000",
			Normalizeduid: 7000,
			Caserecnumber: "545534",
			OnlineLpaId:   "A123",
			Batchid:       "x",
			Casetype:      "lpa",
			Casesubtype:   "hw",
		}, {
			UID:           "7002",
			Normalizeduid: 7002,
			Caserecnumber: "545532",
			OnlineLpaId:   "A124",
			Batchid:       "y",
			Casetype:      "lpa",
			Casesubtype:   "pfa",
		}},
	}, first)

	second, ok := read(resultsCh, time.Second)
	if !assert.True(ok) {
		return
	}
	assert.Equal(Person{
		ID:               i64(2),
		UID:              "7002",
		Normalizeduid:    7002,
		CaseRecNumber:    "",
		DeputyNumber:     nil,
		Email:            "",
		Dob:              "02/01/1990",
		Firstname:        "Jack",
		Middlenames:      "",
		Surname:          "Jackson",
		CompanyName:      "",
		Persontype:       "Donor",
		OrganisationName: "",
		Phonenumbers:     nil,
		Addresses:        nil,
		Cases:            nil,
	}, second)

	_, ok = read(resultsCh, time.Nanosecond)
	assert.False(ok)
}
