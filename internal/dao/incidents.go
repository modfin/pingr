package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
	"time"
)

type IncidentWithTestName struct {
	pingr.Incident
	TestName string `json:"test_name" db:"test_name"`
}

type IncidentContactLogWithName struct {
	pingr.IncidentContactLog
	ContactName string `json:"contact_name" db:"contact_name"`
}

type IncidentFull struct {
	Incident IncidentWithTestName `json:"incident"`
	ContactLog []IncidentContactLogWithName `json:"contact_log"`
}

func GetIncident(id string, db *sqlx.DB) (IncidentFull, error) {
	q := `
	SELECT i.*, t.test_name 
	FROM incidents i
	INNER JOIN tests t on i.test_id = t.test_id
	WHERE i.incident_id = $1
`
	var incident IncidentFull
	err := db.Get(&incident.Incident, q, id)
	if err != nil{
		return incident, err
	}
	q2 := `
	SELECT icl.*, c.contact_name 
	FROM incident_contact_log icl
	INNER JOIN contacts c on icl.contact_id = c.contact_id
	WHERE icl.incident_id = $1
	ORDER BY icl.created_at DESC
`
	err = db.Select(&incident.ContactLog, q2, id)
	return incident, err
}

func GetIncidents(db *sqlx.DB) ([]IncidentWithTestName, error) {
	q := `
	SELECT incident_id,test_id, i.active, i.root_cause, i.created_at, i.closed_at, test_name, i.test_id 
	FROM incidents i
	INNER JOIN tests t USING(test_id)
	ORDER BY i.created_at DESC
`
	var incidents []IncidentWithTestName
	err := db.Select(&incidents, q)

	return incidents, err
}

func GetActiveIncident(testId string, db *sqlx.DB) (pingr.Incident, error) {
	q := `
	SELECT incident_id, test_id, active,root_cause,created_at FROM incidents 
	WHERE active AND test_id = $1
`
	var incident pingr.Incident
	err := db.Get(&incident, q, testId)

	if err != nil {
		return incident, err
	}
	return incident, nil
}

func PostIncident(incident pingr.Incident, db *sqlx.DB) (uint64, error) {
	q := `
	INSERT INTO incidents(test_id, active, root_cause, created_at)  
	VALUES(:test_id, :active, :root_cause, :created_at)
`
	rows, err := db.NamedExec(q, incident)
	if err != nil {
		return 0, err
	}
	incidentId, err := rows.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(incidentId), nil
}

func CloseIncident(incidentId uint64, db *sqlx.DB) error {
	q := `
	UPDATE incidents
	SET active = 0,
	    closed_at = $1
	WHERE incident_id = $2
`
	_, err := db.Exec(q, time.Now(), incidentId)
	if err != nil {
		return err
	}

	return nil
}

func CloseTestIncident(testId string, db *sqlx.DB) error {
	q := `
	UPDATE incidents
	SET active = 0,
	    closed_at = $1
	WHERE test_id = $2
`
	_, err := db.Exec(q, time.Now(), testId)
	if err != nil {
		return err
	}

	return nil
}
