package index

import (
	"context"
	"opg-search-service/person"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
)

func (r *Indexer) queryByID(ctx context.Context, results chan<- person.Person, start, end, batchSize int) error {
	defer func() { close(results) }()

	batch := &batchIter{start: start, end: end, size: batchSize}

	for batch.Next() {
		r.log.Printf("reading range from db (%d, %d)", batch.From(), batch.To())

		rows, err := r.conn.Query(ctx, makeQuery(`p.id >= $1 AND p.id <= $2`), batch.From(), batch.To())
		if err != nil {
			return err
		}

		if err := scan(ctx, rows, results); err != nil {
			return err
		}
	}

	return nil
}

func (r *Indexer) queryByDate(ctx context.Context, results chan<- person.Person, from time.Time) error {
	rows, err := r.conn.Query(ctx, makeQuery(`p.updatedDate >= $1`), from)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

func makeQuery(whereClause string) string {
	return `SELECT p.id, p.uid, coalesce(p.caseRecNumber, ''), coalesce(p.email, ''), coalesce(to_char(p.dob, 'DD/MM/YYYY'), ''),
		coalesce(p.firstname, ''), coalesce(p.middlenames, ''), coalesce(p.surname, ''), coalesce(p.companyname, ''), p.type, coalesce(p.organisationname, ''),
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

func scan(ctx context.Context, rows pgx.Rows, results chan<- person.Person) error {
	var err error
	lastID := -1
	a := &personAdded{}
	var p *person.Person

	for rows.Next() {
		var v rowResult

		err = rows.Scan(&v.ID, &v.UID, &v.CaseRecNumber, &v.Email, &v.Dob,
			&v.Firstname, &v.Middlenames, &v.Surname, &v.CompanyName, &v.Type, &v.OrganisationName,
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
			p = &person.Person{}
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

type rowResult struct {
	ID                 int
	UID                int
	CaseRecNumber      string
	PhoneNumberID      *int
	PhoneNumber        string
	Email              string
	Dob                string
	Firstname          string
	Middlenames        string
	Surname            string
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

func addResultToPerson(a *personAdded, p *person.Person, s rowResult) {
	if p.ID == nil {
		id := int64(s.ID)
		p.ID = &id
		p.UID = strconv.Itoa(s.UID)
		p.Normalizeduid = int64(s.UID)
		p.CaseRecNumber = s.CaseRecNumber
		p.Email = s.Email
		p.Dob = s.Dob
		p.Firstname = s.Firstname
		p.Middlenames = s.Middlenames
		p.Surname = s.Surname
		p.CompanyName = s.CompanyName
		p.Persontype = s.Type
		p.OrganisationName = s.OrganisationName
	}

	if s.AddressID != nil && !a.hasAddress(*s.AddressID) {
		p.Addresses = append(p.Addresses, person.PersonAddress{
			Addresslines: getAddressLines(s.AddressLines),
			Postcode:     s.Postcode,
		})
	}

	if s.PhoneNumberID != nil && !a.hasPhonenumber(*s.PhoneNumberID) {
		p.Phonenumbers = append(p.Phonenumbers, person.PersonPhonenumber{
			Phonenumber: s.PhoneNumber,
		})
	}

	if s.CaseID != nil && !a.hasCase(*s.CaseID) {
		p.Cases = append(p.Cases, person.PersonCase{
			UID:           strconv.Itoa(*s.CasesUID),
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
		r := make([]string, len(v))
		for i, x := range v {
			r[i] = x.(string)
		}
		return r

	case map[string]interface{}:
		r := make([]string, 3)
		for k, x := range v {
			i, err := strconv.Atoi(k)
			if err != nil {
				return nil
			}

			r[i] = x.(string)
		}
		return r

	default:
		return nil
	}
}
