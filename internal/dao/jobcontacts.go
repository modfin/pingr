package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

type JobContactType struct {
	ContactId 	uint64 	`db:"contact_id"`
	ContactType string 	`db:"contact_type"`
	ContactUrl 	string 	`db:"contact_url"`
	Threshold	uint 	`db:"threshold"`
}

func GetAllJobContacts(db *sqlx.DB) ([]pingr.JobContact, error) {
	q := `
		SELECT * FROM job_contacts
	`
	var contacts []pingr.JobContact
	err := db.Select(&contacts, q)
	if err != nil {
		return nil, err
	}

	return contacts, nil
}

func GetJobContacts(id uint64, db *sqlx.DB) (pingr.JobContact, error) {
	q := `
		SELECT * FROM job_contacts 
		WHERE job_id = $1
	`
	var c pingr.JobContact
	err := db.Get(&c, q, id)
	if err != nil {
		return c, nil
	}

	return c, nil
}

func GetJobContactsType(id uint64, db *sqlx.DB) ([]JobContactType, error) {
	q := `
		SELECT c.contact_id, c.contact_type, c.contact_url, jc.threshold FROM job_contacts jc
		INNER JOIN contacts c ON jc.contact_id = c.contact_id
		WHERE jc.job_id = $1
	`
	var jobContactTypes []JobContactType
	err := db.Select(&jobContactTypes, q, id)
	if err != nil {
		return jobContactTypes, nil
	}

	return jobContactTypes, nil
}

func PostJobContact(jobContact pingr.JobContact, db *sqlx.DB) error {
	q := `
		INSERT INTO job_contacts(contact_id, job_id, threshold) 
		VALUES (:contact_id,:job_id,:threshold);
	`
	_, err := db.NamedExec(q, jobContact)
	if err != nil {
		return err
	}

	return nil
}

func DeleteJobContact(jId, cId uint64, db *sqlx.DB) error {
	q := `
		DELETE FROM job_contacts 
		WHERE contact_id = $1 AND job_id = $2
	`
	_, err := db.Exec(q, cId, jId)
	if err != nil {
		return err
	}

	return nil
}

func DeleteJobContacts(jId uint64, db *sqlx.DB) error {
	q := `
		DELETE FROM job_contacts 
		WHERE job_id = $2
	`
	_, err := db.Exec(q, jId)
	if err != nil {
		return err
	}

	return nil
}