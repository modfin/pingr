package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

func GetLogs(db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs
	`
	var logs []pingr.Log

	err := db.Select(&logs, q)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func GetLog(id uint64, db *sqlx.DB) (l pingr.Log, err error) {
	q := `
		SELECT * FROM logs WHERE log_id = $1
	`

	err = db.Get(&l, q, id)
	if err != nil {
		return
	}

	return
}

func PostLog(log pingr.Log, db *sqlx.DB) error {
	q := `
		INSERT INTO logs(job_id, status, message, created_at) 
		VALUES(:job_id,:status,:message,:created_at);
	`
	_, err := db.NamedExec(q, log)
	if err != nil {
		return err
	}

	return nil
}

func DeleteLog(logId uint64, db *sqlx.DB) error {
	q := `
		DELETE FROM logs
		WHERE log_id = $1
	`
	_, err := db.Exec(q, logId)
	if err != nil {
		return err
	}

	return nil
}