package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

func GetContacts(db *sqlx.DB) ([]pingr.Contact, error) {
	q := `
		SELECT * FROM contacts
	`
	var contacts []pingr.Contact
	err := db.Select(&contacts, q)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func GetContact(id string, db *sqlx.DB) (pingr.Contact, error) {
	q := `
		SELECT * FROM contacts 
		WHERE contact_id = $1
	`
	var c pingr.Contact
	err := db.Get(&c, q, id)
	if err != nil {
		return c, err
	}

	return c, nil
}

func PostContact(contact pingr.Contact, db *sqlx.DB) error {
	q := `
		INSERT INTO contacts(contact_id, contact_name, contact_type, contact_url) 
		VALUES (:contact_id,:contact_name,:contact_type,:contact_url);
	`
	_, err := db.NamedExec(q, contact)
	if err != nil {
		return err
	}

	return nil
}

func PutContact(contact pingr.Contact,  db *sqlx.DB)  error {
	q := `
		UPDATE contacts 
		SET contact_name=:contact_name,
			contact_type = :contact_type,
			contact_url = :contact_url
		WHERE contact_id = :contact_id
	`
	_, err := db.NamedExec(q, contact)
	if err != nil {
		return err
	}

	return nil
}

func DeleteContact(id string, db *sqlx.DB) error {
	q := `
		DELETE FROM contacts 
		WHERE contact_id = $1
	`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}