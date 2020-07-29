package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

type FullLog struct {
	pingr.Log
	StatusName string `db:"status_name"`
}

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

func GetTestLogs(id string, db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs 
		WHERE test_id = $1
		ORDER BY created_at DESC 
	`
	var logs []pingr.Log

	err := db.Select(&logs, q, id)
	if err != nil {
		return nil, err
	}

	return logs, nil
}


func GetTestLogsLimited(id string, limit int, db *sqlx.DB) ([]FullLog, error) {
	q := `
		SELECT sm.status_name, message, created_at, response_time FROM logs
		INNER JOIN status_map sm on logs.status_id = sm.status_id
		WHERE test_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	var logs []FullLog

	err := db.Select(&logs, q, id, limit)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func GetTestLogsDaysLimited(id string, days string, db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs
		WHERE test_id = $1
		AND created_at > datetime('now', '-'||$2||' days')
		ORDER BY created_at DESC
	`
	var logs []pingr.Log

	err := db.Select(&logs, q, id, days)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func PostLog(log pingr.Log, db *sqlx.DB) error {
	q := `
		INSERT INTO logs(test_id, status_id, message, response_time, created_at) 
		VALUES(:test_id,:status_id,:message,:response_time,:created_at);
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

func DeleteTestLogs(testId string, db *sqlx.DB) error {
	q := `
		DELETE FROM logs
		WHERE test_id = $1
	`
	_, err := db.Exec(q, testId)
	if err != nil {
		return err
	}

	return nil
}

func DeleteLastNLogs(n uint, db *sqlx.DB) error {
	q := `
		DELETE FROM logs
		WHERE log_id IN 
			(SELECT log_id FROM logs ORDER BY created_at LIMIT $1)
	`
	_, err := db.Exec(q, n)
	if err != nil {
		return err
	}

	return nil
}