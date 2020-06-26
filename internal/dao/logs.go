package dao

import (
	"database/sql"
)

func GetLogs(db *sql.DB) ([]Log, error) {
	q := `
		SELECT * FROM logs
	`
	rows, err := db.Query(q)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var logs []Log
	for rows.Next() {
		var l Log
		err := rows.Scan(&l.LogId, &l.JobId, &l.Status, &l.Message, &l.CreatedAt)
		if err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func GetLog(id uint64, db *sql.DB) (l Log, err error) {
	q := `
		SELECT * FROM logs WHERE log_id = ?
	`
	stmt, err := db.Prepare(q)
	defer stmt.Close()
	if err != nil {
		return
	}

	err = stmt.QueryRow(id).Scan(&l.LogId, &l.JobId, &l.Status, &l.Message, &l.CreatedAt)
	if err != nil {
		return
	}
	return
}

func PostLog(log Log, db *sql.DB) error {
	q := `
		INSERT INTO logs(job_id, status, message, created_at) VALUES(?,?,?,?);
	`
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	defer stmt.Close()
	if err != nil {
		return err
	}


	_, err = stmt.Exec(log.JobId, log.Status, log.Message, log.CreatedAt)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func DeleteLog(logId uint64, db *sql.DB) error {
	q := `
		DELETE FROM logs
		WHERE log_id = ?
	`
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	defer stmt.Close()
	if err != nil {
		return err
	}

	_, err = stmt.Exec(logId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}