package index

import (
	"context"
	"opg-search-service/person"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct{}

func (*mockLogger) Printf(s string, args ...interface{}) {}

func TestGetIDRange(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://searchservice:searchservice@postgres:5432/searchservice")
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("../../../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	_, err = conn.Exec(ctx, `
INSERT INTO persons (id, uid, caserecnumber, email, dob, firstname, middlenames, surname, companyname, type, organisationname)
						 VALUES (1, 7000, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'lpa_donor', 'Orgz'),
										(2, 7002, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'lpa_donor', null),
										(3, 7003, null, null, '1990-01-02', 'J', null, 'J', null, 'lpa_donor', null);
`)
	if !assert.Nil(err) {
		return
	}

	r := &Indexer{conn: conn, log: &mockLogger{}}

	min, max, err := r.getIDRange(ctx)
	assert.Nil(err)
	assert.Equal(1, min)
	assert.Equal(3, max)
}

func TestQueryByID(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://searchservice:searchservice@postgres:5432/searchservice")
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("../../../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	_, err = conn.Exec(ctx, `
INSERT INTO persons (id, uid, caserecnumber, email, dob, firstname, middlenames, surname, companyname, type, organisationname)
						 VALUES (1, 7000, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'lpa_donor', 'Orgz'),
										(2, 7002, null, null, '1990-01-02', 'Jack', null, 'Jackson', null, 'lpa_donor', null),
										(3, 7003, null, null, '1990-01-02', 'J', null, 'J', null, 'lpa_donor', null);

INSERT INTO phonenumbers (id, person_id, phone_number)
									VALUES (1, 1, '077777777');

INSERT INTO addresses (id, person_id, address_lines, postcode)
							 VALUES (1, 1, json_build_object('0', '1 Road', '2', 'Place'), 'S1 1AB');

INSERT INTO cases (id, uid, caserecnumber, onlinelpaid, batchid, casetype, casesubtype)
					 VALUES (1, 7000, '545534', 'A123', 'x', 'lpa', 'hw'),
									(2, 7002, '545532', 'A124', 'y', 'lpa', 'pfa');

INSERT INTO person_caseitem (person_id, caseitem_id) VALUES (1, 1), (1, 2);
`)
	if !assert.Nil(err) {
		return
	}

	r := &Indexer{conn: conn, log: &mockLogger{}}

	resultsCh := make(chan person.Person)
	results := []person.Person{}
	go func() {
		for p := range resultsCh {
			results = append(results, p)
		}
	}()

	err = r.queryByID(ctx, resultsCh, 1, 2, 10)
	if assert.Nil(err) && assert.Len(results, 2) {
		assert.Equal([]person.Person{{
			ID:               i64(1),
			UID:              "7000",
			Normalizeduid:    7000,
			CaseRecNumber:    "1010101",
			Email:            "email@example.com",
			Dob:              "02/01/2002",
			Firstname:        "John",
			Middlenames:      "J",
			Surname:          "Johnson",
			CompanyName:      "& co",
			Persontype:       "lpa_donor",
			OrganisationName: "Orgz",
			Phonenumbers: []person.PersonPhonenumber{{
				Phonenumber: "077777777",
			}},
			Addresses: []person.PersonAddress{{
				Addresslines: []string{"1 Road", "", "Place"},
				Postcode:     "S1 1AB",
			}},
			Cases: []person.PersonCase{{
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
		}, {
			ID:               i64(2),
			UID:              "7002",
			Normalizeduid:    7002,
			CaseRecNumber:    "",
			Email:            "",
			Dob:              "02/01/1990",
			Firstname:        "Jack",
			Middlenames:      "",
			Surname:          "Jackson",
			CompanyName:      "",
			Persontype:       "lpa_donor",
			OrganisationName: "",
			Phonenumbers:     nil,
			Addresses:        nil,
			Cases:            nil,
		}}, results)
	}
}

func TestQueryByDate(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://searchservice:searchservice@postgres:5432/searchservice")
	if !assert.Nil(err) {
		return
	}
	defer conn.Close(ctx)

	schemaSql, _ := os.ReadFile("../../../testdata/schema.sql")

	_, err = conn.Exec(ctx, string(schemaSql))
	if !assert.Nil(err) {
		return
	}

	_, err = conn.Exec(ctx, `
INSERT INTO persons (id, updateddate, uid, caserecnumber, email, dob, firstname, middlenames, surname, companyname, type, organisationname)
						 VALUES (1, '2021-01-03 12:00:00', 7000, 1010101, 'email@example.com', '2002-01-02', 'John', 'J', 'Johnson', '& co', 'lpa_donor', 'Orgz'),
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

	r := &Indexer{conn: conn}

	resultsCh := make(chan person.Person)
	results := []person.Person{}
	go func() {
		for p := range resultsCh {
			results = append(results, p)
		}
	}()

	err = r.queryByDate(ctx, resultsCh, time.Date(2021, time.January, 1, 23, 0, 0, 0, time.UTC))
	if assert.Nil(err) && assert.Len(results, 2) {
		assert.Equal([]person.Person{{
			ID:               i64(1),
			UID:              "7000",
			Normalizeduid:    7000,
			CaseRecNumber:    "1010101",
			Email:            "email@example.com",
			Dob:              "02/01/2002",
			Firstname:        "John",
			Middlenames:      "J",
			Surname:          "Johnson",
			CompanyName:      "& co",
			Persontype:       "lpa_donor",
			OrganisationName: "Orgz",
			Phonenumbers: []person.PersonPhonenumber{{
				Phonenumber: "077777777",
			}},
			Addresses: []person.PersonAddress{{
				Addresslines: []string{"123 Fake St"},
				Postcode:     "S1 1AB",
			}},
			Cases: []person.PersonCase{{
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
		}, {
			ID:               i64(2),
			UID:              "7002",
			Normalizeduid:    7002,
			CaseRecNumber:    "",
			Email:            "",
			Dob:              "02/01/1990",
			Firstname:        "Jack",
			Middlenames:      "",
			Surname:          "Jackson",
			CompanyName:      "",
			Persontype:       "lpa_donor",
			OrganisationName: "",
			Phonenumbers:     nil,
			Addresses:        nil,
			Cases:            nil,
		}}, results)
	}
}

func i64(x int) *int64 {
	y := int64(x)
	return &y
}
