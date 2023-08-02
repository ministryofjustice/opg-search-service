package poadraftapplication

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/index"
)

type DB struct {
	conn *pgx.Conn
}

type rowResult struct {
	ID int
	UID string
	donorname string
	donorpostcode string
	donordob string
}

func NewDB(conn *pgx.Conn) *DB {
	return &DB{conn: conn}
}

func (db *DB) QueryIDRange(ctx context.Context) (min int, max int, err error) {
	return min, max, db.conn.QueryRow(ctx, "SELECT COALESCE(MIN(id), 0), COALESCE(MAX(id), 0) FROM poa.draft_applications").Scan(&min, &max)
}

func (db *DB) QueryByID(ctx context.Context, results chan<- index.Indexable, from, to int) error {
	query := `SELECT COALESCE(uid, ''), COALESCE(donorname, ''), COALESCE(donordob, ''), COALESCE(donorpostcode, '')
FROM poa.draft_applications
WHERE id >= $1 AND id <= $2
ORDER BY id`

	rows, err := db.conn.Query(ctx, query, from, to)
	if err != nil {
		return err
	}

	for rows.Next() {
		var r rowResult
		err = rows.Scan(&r.UID, &r.donorname, &r.donordob, &r.donorpostcode)

		if err != nil {
			break
		}

		d := DraftApplication{
			UID: &r.UID,
			Donor: DraftApplicationDonor{
				Name: r.donorname,
				Dob: r.donordob,
				Postcode: r.donorpostcode,
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
