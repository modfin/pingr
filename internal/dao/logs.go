package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

func GetLogs(db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs
		ORDER BY created_at DESC 
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
		SELECT * FROM logs 
		WHERE log_id = $1
		ORDER BY created_at DESC 
	`

	err = db.Get(&l, q, id)
	if err != nil {
		return
	}

	return
}

func GetJobLogs(id uint64, db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs 
		WHERE job_id = $1
		ORDER BY created_at DESC 
	`
	var logs []pingr.Log

	err := db.Select(&logs, q, id)
	if err != nil {
		return nil, err
	}

	return logs, nil
}


func GetJobLogsLimited(id uint64, limit int, db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs
		WHERE job_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	var logs []pingr.Log

	err := db.Select(&logs, q, id, limit)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func PostLog(log pingr.Log, db *sqlx.DB) error {
	q := `
		INSERT INTO logs(job_id, status_id, message, response_time, created_at) 
		VALUES(:job_id,:status_id,:message,:response_time,:created_at);
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