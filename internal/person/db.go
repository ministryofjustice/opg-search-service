package person

import (
	"context"
	"fmt"
	"sort"
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
	err = db.conn.QueryRow(ctx, "SELECT COALESCE(MIN(id), 0), COALESCE(MAX(id), 0) FROM persons").Scan(&min, &max)

	return min, max, err
}

func (db *DB) QueryByID(ctx context.Context, results chan<- index.Indexable, from, to int) error {
	rows, err := db.conn.Query(ctx, makeQueryPerson(`p.id >= $1 AND p.id <= $2`), from, to)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

func (db *DB) QueryFromDate(ctx context.Context, results chan<- index.Indexable, from time.Time) error {
	rows, err := db.conn.Query(ctx, makeQueryPerson(`p.updatedDate >= $1`), from)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

func makeQueryPerson(whereClause string) string {
	return `SELECT p.id, p.uid, coalesce(p.caseRecNumber, ''), p.deputynumber, coalesce(p.email, ''), coalesce(to_char(p.dob, 'DD/MM/YYYY'), ''),
		coalesce(p.firstname, ''), coalesce(p.middlenames, ''), coalesce(p.surname, ''), coalesce(p.previousnames, ''), coalesce(p.othernames, ''), coalesce(p.companyname, ''), p.type, coalesce(p.organisationname, ''),
		phonenumbers.id, coalesce(phonenumbers.phone_number, ''),
		addresses.id, addresses.address_lines, coalesce(addresses.postcode, ''),
		cases.id, cases.uid, coalesce(cases.caserecnumber, ''), coalesce(cases.onlinelpaid, ''), coalesce(cases.batchid, ''), coalesce(cases.casetype, ''), coalesce(cases.casesubtype, '')
FROM persons p
LEFT JOIN phonenumbers ON p.id = phonenumbers.person_id
LEFT JOIN addresses ON p.id = addresses.person_id
LEFT JOIN person_caseitem ON p.id = person_caseitem.person_id
LEFT JOIN cases ON person_caseitem.caseitem_id = cases.id
WHERE ` + whereClause + `
ORDER BY p.id`
}

type rowResult struct {
	ID                 int
	UID                int
	CaseRecNumber      string
	DeputyNumber       *int
	PhoneNumberID      *int
	PhoneNumber        string
	Email              string
	Dob                string
	Firstname          string
	Middlenames        string
	Surname            string
	Previousnames      string
	Othernames         string
	CompanyName        string
	Type               string
	OrganisationName   string
	AddressID          *int
	AddressLines       interface{}
	Postcode           string
	CaseID             *int
	CasesUID           *int
	CasesCaseRecNumber string
	CasesOnlineLpaID   string
	CasesBatchID       string
	CasesCaseType      string
	CasesCaseSubType   string
}

func scan(ctx context.Context, rows pgx.Rows, results chan<- index.Indexable) error {
	var err error
	lastID := -1

	a := &personAdded{}
	var p *Person

	for rows.Next() {
		var v rowResult
		err = rows.Scan(&v.ID, &v.UID, &v.CaseRecNumber, &v.DeputyNumber, &v.Email, &v.Dob,
			&v.Firstname, &v.Middlenames, &v.Surname, &v.Previousnames, &v.Othernames, &v.CompanyName, &v.Type, &v.OrganisationName,
			&v.PhoneNumberID, &v.PhoneNumber,
			&v.AddressID, &v.AddressLines, &v.Postcode,
			&v.CaseID, &v.CasesUID, &v.CasesCaseRecNumber, &v.CasesOnlineLpaID, &v.CasesBatchID, &v.CasesCaseType, &v.CasesCaseSubType)

		if err != nil {
			break
		}

		if v.ID != lastID {
			if p != nil {
				results <- *p
			}

			a.clear()
			p = &Person{}
			lastID = v.ID
		}
		addResultToPerson(a, p, v)
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

type personAdded struct {
	addresses    map[int]struct{}
	phonenumbers map[int]struct{}
	cases        map[int]struct{}
}

func (a *personAdded) hasAddress(id int) bool {
	_, ok := a.addresses[id]
	a.addresses[id] = struct{}{}
	return ok
}

func (a *personAdded) hasPhonenumber(id int) bool {
	_, ok := a.phonenumbers[id]
	a.phonenumbers[id] = struct{}{}
	return ok
}

func (a *personAdded) hasCase(id int) bool {
	_, ok := a.cases[id]
	a.cases[id] = struct{}{}
	return ok
}

func (a *personAdded) clear() {
	a.addresses = map[int]struct{}{}
	a.phonenumbers = map[int]struct{}{}
	a.cases = map[int]struct{}{}
}

func addResultToPerson(a *personAdded, p *Person, s rowResult) {
	if p.ID == nil {
		p.ID = i64(s.ID)
		p.UID = formatUID(s.UID)
		p.Normalizeduid = int64(s.UID)
		p.CaseRecNumber = s.CaseRecNumber
		p.DeputyNumber = nil
		if s.DeputyNumber != nil {
			p.DeputyNumber = i64(*s.DeputyNumber)
		}
		p.Email = s.Email
		p.Dob = s.Dob
		p.Firstname = s.Firstname
		p.Middlenames = s.Middlenames
		p.Surname = s.Surname
		p.Previousnames = s.Previousnames
		p.Othernames = s.Othernames
		p.CompanyName = s.CompanyName
		p.Persontype = resolvePersonType(s.Type)
		p.OrganisationName = s.OrganisationName
	}

	if s.AddressID != nil && !a.hasAddress(*s.AddressID) {
		p.Addresses = append(p.Addresses, PersonAddress{
			Addresslines: getAddressLines(s.AddressLines),
			Postcode:     s.Postcode,
		})
	}

	if s.PhoneNumberID != nil && !a.hasPhonenumber(*s.PhoneNumberID) {
		p.Phonenumbers = append(p.Phonenumbers, PersonPhonenumber{
			Phonenumber: s.PhoneNumber,
		})
	}

	if s.CaseID != nil && !a.hasCase(*s.CaseID) {
		p.Cases = append(p.Cases, PersonCase{
			UID:           formatUID(*s.CasesUID),
			Normalizeduid: int64(*s.CasesUID),
			Caserecnumber: s.CasesCaseRecNumber,
			OnlineLpaId:   s.CasesOnlineLpaID,
			Batchid:       s.CasesBatchID,
			Casetype:      s.CasesCaseType,
			Casesubtype:   s.CasesCaseSubType,
		})
	}
}

func getAddressLines(lines interface{}) []string {
	switch v := lines.(type) {
	case []interface{}:
		r := []string{}
		for _, x := range v {
			if x != nil && x != "" {
				r = append(r, x.(string))
			}
		}
		return r
	case map[string]interface{}:
		keys := make([]string, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		sorted := []string{}
		for _, key := range keys {
			// ignore non-numeric keys
			_, err := strconv.Atoi(key)
			if err != nil {
				continue
			}

			// ignore keys with null values
			value := v[key]
			if value == nil {
				continue
			}

			strValue := value.(string)

			// ignore empty strings
			if strValue == "" {
				continue
			}

			sorted = append(sorted, strValue)
		}

		return sorted
	default:
		return nil

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
	case "lpa_attorney":
		return "Attorney"
	case "lpa_replacement_attorney":
		return "Replacement Attorney"
	case "lpa_trust_corporation":
		return "Trust Corporation"
	case "lpa_correspondent":
		return "Correspondent"
	case "lpa_donor":
		return "Donor"
	case "lpa_notified_person":
		return "Notified Person"
	case "lpa_certificate_provider":
		return "Certificate Provider"
	case "actor_non_case_contact":
		return "Non-Case Contact"
	case "actor_notified_relative":
		return "Notified Relative"
	case "actor_notified_attorney":
		return "Notified Attorney"
	case "actor_notified_donor":
		return "Person Notify Donor"
	case "actor_client":
		return "Client"
	case "actor_contact":
		return "Contact"
	case "actor_deputy":
		return "Deputy"
	case "actor_fee_payer":
		return "Fee Payer"
	}

	return t
}

func i64(x int) *int64 {
	y := int64(x)
	return &y
}
