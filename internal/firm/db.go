package firm

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/index"
)

func NewDB(conn *pgx.Conn) *DB {
	return &DB{conn: conn}
}

type DB struct {
	conn *pgx.Conn
}

func (db *DB) QueryIDRange(ctx context.Context) (min int, max int, err error) {
	err = db.conn.QueryRow(ctx, "SELECT COALESCE(MIN(id), 0), COALESCE(MAX(id), 0) FROM firm").Scan(&min, &max)

	return min, max, err
}

func (db *DB) QueryByID(ctx context.Context, results chan<- index.Indexable, from, to int) error {
	rows, err := db.conn.Query(ctx, makeQueryFirm(`f.id >= $1 AND f.id <= $2`), from, to)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

func (db *DB) QueryFromDate(ctx context.Context, results chan<- index.Indexable, fromDate time.Time) error {
	return errors.New("firms cannot be queried by date")
}

func makeQueryFirm(whereClause string) string {
	return `SELECT f.id, coalesce(f.email, ''), f.firmname, f.firmNumber,
		coalesce(f.addressline1, ''), coalesce(f.addressline2, ''), coalesce(f.addressline3, ''),
		coalesce(f.town, ''), coalesce(f.county, ''), coalesce(f.postcode, ''),
		coalesce(f.phonenumber, '')
FROM firm f
WHERE ` + whereClause + `
ORDER BY f.id`
}

func scan(ctx context.Context, rows pgx.Rows, results chan<- index.Indexable) error {
	var err error
	lastID := -1

	var f *Firm
	for rows.Next() {
		var v rowResultFirm
		err = rows.Scan(&v.ID, &v.Email, &v.FirmName,
			&v.FirmNumber, &v.AddressLine1, &v.AddressLine2, &v.AddressLine3, &v.Town, &v.County,
			&v.Postcode, &v.PhoneNumber)

		if err != nil {
			break
		}

		if v.ID != lastID {
			if f != nil {
				results <- *f
			}

			f = &Firm{}
			lastID = v.ID
		}

		addResultToFirm(f, v)
	}

	if f != nil {
		results <- *f
	}

	if err != nil {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return err

}

type rowResultFirm struct {
	ID           int
	Email        string
	FirmName     string
	FirmNumber   int
	AddressLine1 string
	AddressLine2 string
	AddressLine3 string
	Town         string
	County       string
	Postcode     string
	PhoneNumber  string
}

func addResultToFirm(f *Firm, s rowResultFirm) {
	if f.ID == nil {
		id := int64(s.ID)
		f.ID = &id
		f.Persontype = "Firm"
		f.Email = s.Email
		f.FirmName = s.FirmName
		f.FirmNumber = strconv.Itoa(s.FirmNumber)
		f.AddressLine1 = s.AddressLine1
		f.AddressLine2 = s.AddressLine2
		f.AddressLine3 = s.AddressLine3
		f.Town = s.Town
		f.County = s.County
		f.Postcode = s.Postcode
		f.Phonenumber = s.PhoneNumber
	}
}
