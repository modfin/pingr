package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

type TestContactType struct {
	ContactId   string `json:"contact_id" db:"contact_id"`
	ContactName string `json:"contact_name" db:"contact_name"`
	ContactType string `json:"contact_type" db:"contact_type"`
	ContactUrl  string `json:"contact_url" db:"contact_url"`
	Threshold   uint   `json:"threshold" db:"threshold"`
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

func GetTestContacts(id string, db *sqlx.DB) ([]pingr.TestContact, error) {
	q := `
		SELECT * FROM test_contacts 
		WHERE test_id = $1
	`
	var c []pingr.TestContact
	err := db.Select(&c, q, id)
	if err != nil {
		return c, err
	}

	return c, nil
}

func GetTestContactsType(id string, db *sqlx.DB) ([]TestContactType, error) {
	q := `
		SELECT c.contact_id, c.contact_name, c.contact_type, c.contact_url, jc.threshold FROM test_contacts jc
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

func GetTestContactsToContact(id string, db *sqlx.DB) ([]pingr.Contact, error) {
	q := `
		WITH _test AS (
    SELECT $1 as test_id
),
     _last_log AS (
         SELECT *
         FROM logs
                  INNER JOIN _test
                             USING (test_id)
         WHERE NOT (status_id = 2 OR status_id = 3)
         ORDER BY created_at DESC
         LIMIT 1
     ),
     _failing_test AS (
         SELECT test_id, count(*) fails
         FROM logs
                  INNER JOIN _test
                             USING (test_id)
         WHERE (status_id = 2
             OR status_id = 3)
           AND created_at > (SELECT created_at FROM _last_log)
         GROUP BY test_id
     ),
     _contacts_notified AS (
         SELECT test_id, icl.contact_id
         FROM incidents
                  INNER JOIN incident_contact_log icl
                             USING (incident_id)
         WHERE active
     )
SELECT c.contact_id, contact_name, contact_type, contact_url
FROM test_contacts c
         INNER JOIN _failing_test f
                    USING (test_id)
         INNER JOIN contacts
                    USING (contact_id)
WHERE f.fails >= threshold
  AND (test_id, contact_id) NOT IN _contacts_notified
;
	`
	var contacts []pingr.Contact
	err := db.Select(&contacts, q, id)
	if err != nil {
		return contacts, err
	}

	return contacts, nil
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
