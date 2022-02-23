package index

import (
	"context"
	"fmt"
	"github.com/ministryofjustice/opg-search-service/internal/firm"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/ministryofjustice/opg-search-service/internal/person"
	"github.com/sirupsen/logrus"
)

func (r *Indexer) getIDRangePerson(ctx context.Context) (min int, max int, err error) {
	err = r.conn.QueryRow(ctx, "SELECT MIN(id), MAX(id) FROM persons").Scan(&min, &max)

	return min, max, err
}

func (r *Indexer) getIDRangeFirm(ctx context.Context) (min int, max int, err error) {
	l := logrus.New()
	l.SetFormatter(&logrus.JSONFormatter{})
	l.Println("in get range firm function")

	err = r.conn.QueryRow(ctx, "SELECT MIN(id), MAX(id) FROM firm").Scan(&min, &max)
	l.Println(min, max, err)

	return min, max, err
}

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


func (r *Indexer) queryByIDFirm(ctx context.Context, results chan<- firm.Firm, start, end, batchSize int) error {
	defer func() { close(results) }()

	batch := &batchIter{start: start, end: end, size: batchSize}

	for batch.Next() {
		r.log.Printf("reading range from db (%d, %d)", batch.From(), batch.To())

		rows, err := r.conn.Query(ctx, makeQueryFirm(`f.id >= $1 AND f.id <= $2`), batch.From(), batch.To())
		if err != nil {
			return err
		}

		if err := scanFirm(ctx, rows, results); err != nil {
			return err
		}
	}

	return nil
}


func (r *Indexer) queryFromDate(ctx context.Context, results chan<- person.Person, from time.Time) error {
	defer func() { close(results) }()

	rows, err := r.conn.Query(ctx, makeQuery(`p.updatedDate >= $1`), from)
	if err != nil {
		return err
	}

	return scan(ctx, rows, results)
}

//func (r *Indexer) queryFromDateFirm(ctx context.Context, results chan<- person.Person, from time.Time) error {
//	defer func() { close(results) }()
//
//	rows, err := r.conn.Query(ctx, makeQuery(`f.updatedDate >= $1`), from)
//	if err != nil {
//		return err
//	}
//
//	return scan(ctx, rows, results)
//}


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

func makeQueryFirm(whereClause string) string {
	return `SELECT f.id, coalesce(f.email, ''), f.firmname, f.firmNumber,
		coalesce(f.addressline1, ''), coalesce(f.addressline2, ''), coalesce(f.addressline3, ''), 
		coalesce(f.town, ''), coalesce(f.county, ''), coalesce(f.postcode, ''),
		coalesce(f.phonenumber, '')
FROM firm f
WHERE ` + whereClause + `
ORDER BY f.id`
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


func scanFirm(ctx context.Context, rows pgx.Rows, results chan<- firm.Firm) error {
	var err error
	lastID := -1
	var f *firm.Firm

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

			f = &firm.Firm{}
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
		p.UID = formatUID(s.UID)
		p.Normalizeduid = int64(s.UID)
		p.CaseRecNumber = s.CaseRecNumber
		p.Email = s.Email
		p.Dob = s.Dob
		p.Firstname = s.Firstname
		p.Middlenames = s.Middlenames
		p.Surname = s.Surname
		p.CompanyName = s.CompanyName
		p.Persontype = resolvePersonType(s.Type)
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


func addResultToFirm(f *firm.Firm, s rowResultFirm) {
	if f.ID == nil {
		id := int64(s.ID)
		f.ID = &id
		f.Email = s.Email
		f.FirmName = s.FirmName
		f.FirmNumber = s.FirmNumber
		f.AddressLine1 = s.AddressLine1
		f.AddressLine2 = s.AddressLine2
		f.AddressLine3 = s.AddressLine3
		f.Town = s.Town
		f.County = s.County
		f.Postcode = s.Postcode
		f.Phonenumber = s.PhoneNumber
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
