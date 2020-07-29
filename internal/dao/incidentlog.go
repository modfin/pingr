package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

func PostContactLog(log pingr.IncidentContactLog, db *sqlx.DB) error {
	q := `
	INSERT INTO incident_contact_log(incident_id, contact_id, message, created_at)  
	VALUES(:incident_id, :contact_id, :message, :created_at)
`

	_, err := db.NamedExec(q, log)
	if err != nil {
		return err
	}

	return nil
}

func GetIncidentContacts(incidentId uint64, db *sqlx.DB) ([]pingr.Contact, error) {
	q := `
	SELECT c.contact_id, contact_name, c.contact_type, c.contact_url 
	FROM incident_contact_log i
	INNER JOIN contacts c on i.contact_id = c.contact_id
	WHERE i.incident_id = $1
	`

	var contacts []pingr.Contact
	err := db.Select(&contacts, q, incidentId)
	if err != nil {
		return contacts, err
	}

	return contacts, nil
}
