package poadraftapplication

import (
	"context"
	"errors"
	"time"

	"github.com/ministryofjustice/opg-search-service/internal/index"
)

type DB struct {
	conn *pgx.Conn
}

type rowResult struct {
	ID              int
	UID             string
	donorfirstnames string
	donorlastname   string
	donorpostcode   string
	donordob        string
}

func NewDB(conn *pgx.Conn) *DB {
	return &DB{conn: conn}
}

func (db *DB) QueryIDRange(ctx context.Context) (min int, max int, err error) {
	return min, max, db.conn.QueryRow(ctx, "SELECT COALESCE(MIN(id), 0), COALESCE(MAX(id), 0) FROM poa.draft_applications").Scan(&min, &max)
}

func (db *DB) QueryByID(ctx context.Context, results chan<- index.Indexable, from, to int) error {
	query := `SELECT COALESCE(c.caserecnumber, ''), COALESCE(a.donorfirstnames, ''), COALESCE(a.donorlastname, ''), coalesce(to_char(a.donordob, 'DD/MM/YYYY'), ''), COALESCE(a.donorpostcode, '')
FROM poa.draft_applications a
INNER JOIN cases c ON c.id = a.lpa_id
WHERE a.id >= $1 AND a.id <= $2
ORDER BY a.id`

	rows, err := db.conn.Query(ctx, query, from, to)
	if err != nil {
		return err
	}

	for rows.Next() {
		var r rowResult
		err = rows.Scan(&r.UID, &r.donorfirstnames, &r.donorlastname, &r.donordob, &r.donorpostcode)

		if err != nil {
			break
		}

		d := DraftApplication{
			UID: r.UID,
			Donor: DraftApplicationDonor{
				FirstNames: r.donorfirstnames,
				LastName:   r.donorlastname,
				Dob:        r.donordob,
				Postcode:   r.donorpostcode,
			},
		}

		results <- d
	}

	if err != nil {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return err
}

func (db *DB) QueryFromDate(ctx context.Context, results chan<- index.Indexable, fromDate time.Time) error {
	return errors.New("draft applications cannot be queried by date")
}
