package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

type TestContactType struct {
	ContactId 	string 	`db:"contact_id"`
	ContactType string 	`db:"contact_type"`
	ContactUrl 	string 	`db:"contact_url"`
	Threshold	uint 	`db:"threshold"`
}

func GetAllTestContacts(db *sqlx.DB) ([]pingr.TestContact, error) {
	q := `
		SELECT * FROM test_contacts
	`
	var contacts []pingr.TestContact
	err := db.Select(&contacts, q)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func GetTestContacts(id string, db *sqlx.DB) (pingr.TestContact, error) {
	q := `
		SELECT * FROM test_contacts 
		WHERE test_id = $1
	`
	var c pingr.TestContact
	err := db.Get(&c, q, id)
	if err != nil {
		return c, err
	}

	return c, nil
}

func GetTestContactsType(id string, db *sqlx.DB) ([]TestContactType, error) {
	q := `
		SELECT c.contact_id, c.contact_type, c.contact_url, jc.threshold FROM test_contacts jc
		INNER JOIN contacts c ON jc.contact_id = c.contact_id
		WHERE jc.test_id = $1
	`
	var testContactTypes []TestContactType
	err := db.Select(&testContactTypes, q, id)
	if err != nil {
		return testContactTypes, err
	}

	return testContactTypes, nil
}

func PostTestContact(testContact pingr.TestContact, db *sqlx.DB) error {
	q := `
		INSERT INTO test_contacts(contact_id, test_id, threshold) 
		VALUES (:contact_id,:test_id,:threshold);
	`
	_, err := db.NamedExec(q, testContact)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTestContact(jId, cId string, db *sqlx.DB) error {
	q := `
		DELETE FROM test_contacts 
		WHERE contact_id = $1 AND test_id = $2
	`
	_, err := db.Exec(q, cId, jId)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTestContacts(jId string, db *sqlx.DB) error {
	q := `
		DELETE FROM test_contacts 
		WHERE test_id = $2
	`
	_, err := db.Exec(q, jId)
	if err != nil {
		return err
	}

	return nil
}