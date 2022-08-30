package deputy

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/index"
	"strconv"
	"time"
)

func NewDB(conn *pgx.Conn) *DB {
	return &DB{conn: conn}
}

type DB struct {
	conn *pgx.Conn
}

func (db *DB) QueryIDRange(ctx context.Context) (min int, max int, err error) {
	err = db.conn.QueryRow(ctx, "SELECT MIN(id), MAX(id) FROM persons").Scan(&min, &max)

	return min, max, err
}

func (db *DB) QueryByID(ctx context.Context, results chan<- index.Indexable, from, to int) error {
	rows, err := db.conn.Query(ctx, makeQueryDeputy(`p.id >= $1 AND p.id <= $2`), from, to)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

func (db *DB) QueryFromDate(ctx context.Context, results chan<- index.Indexable, from time.Time) error {
	rows, err := db.conn.Query(ctx, makeQueryDeputy(`p.updatedDate >= $1`), from)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

func makeQueryDeputy(whereClause string) string {
	return `SELECT p.id, p.uid, p.deputynumber, coalesce(to_char(p.dob, 'DD/MM/YYYY'), ''),
		coalesce(p.firstname, ''), coalesce(p.middlenames, ''), coalesce(p.surname, ''), coalesce(p.companyname, ''), p.type, coalesce(p.organisationname, '')
FROM persons p
WHERE ` + whereClause + `
ORDER BY p.id`
}

type rowResult struct {
	ID               int
	UID              int
	DeputyNumber     *int
	Dob              string
	Firstname        string
	Middlenames      string
	Surname          string
	CompanyName      string
	Type             string
	OrganisationName string
}

func scan(ctx context.Context, rows pgx.Rows, results chan<- index.Indexable) error {
	var err error
	lastID := -1

	var p *Deputy

	for rows.Next() {
		var v rowResult
		err = rows.Scan(&v.ID, &v.UID, &v.DeputyNumber, &v.Dob,
			&v.Firstname, &v.Middlenames, &v.Surname, &v.CompanyName, &v.Type, &v.OrganisationName)

		if err != nil {
			break
		}

		if v.ID != lastID {
			if p != nil {
				results <- *p
			}

			p = &Deputy{}
			lastID = v.ID
		}
		addResultToDeputy(p, v)
	}

	if p != nil {
		results <- *p
	}

	if err != nil {
		return err
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return err

}

func addResultToDeputy(p *Deputy, s rowResult) {
	if p.ID == nil {
		p.ID = i64(s.ID)
		p.UID = formatUID(s.UID)
		p.Normalizeduid = int64(s.UID)
		p.DeputyNumber = nil
		if s.DeputyNumber != nil {
			p.DeputyNumber = i64(*s.DeputyNumber)
		}
		p.Dob = s.Dob
		p.Firstname = s.Firstname
		p.Middlenames = s.Middlenames
		p.Surname = s.Surname
		p.CompanyName = s.CompanyName
		p.Persontype = resolvePersonType(s.Type)
		p.OrganisationName = s.OrganisationName
	}
}

func formatUID(uid int) string {
	s := strconv.Itoa(uid)
	if len(s) != 12 {
		return s
	}

	return fmt.Sprintf("%s-%s-%s", s[0:4], s[4:8], s[8:12])
}

func resolvePersonType(t string) string {
	switch t {
	case "actor_deputy":
		return "Deputy"
	}

	return t
}

func i64(x int) *int64 {
	y := int64(x)
	return &y
}
