package dao

import (
	"database/sql"
)

func GetJobs(db *sql.DB) ([]Job, error) {
	q := `
		SELECT * FROM jobs
	`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		err := rows.Scan(&j.JobId, &j.TestType, &j.Url, &j.Interval, &j.Timeout, &j.CreatedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func GetJob(id string, db *sql.DB) (j Job, err error) {
	q := `
		SELECT * FROM jobs WHERE job_id = ?
	`
	stmt, err := db.Prepare(q)
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&j.JobId, &j.TestType, &j.Url, &j.Interval, &j.Timeout, &j.CreatedAt)
	if err != nil {
		return
	}
	return
}

func GetJobLogs(id string, db *sql.DB) ([]Log, error) {
	q := `
		SELECT * FROM logs WHERE job_id = ?
	`

	stmt, err := db.Prepare(q)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func PostJob(job Job, db *sql.DB) error {
	q := `
		INSERT INTO jobs(test_type, url, interval, timeout, created_at) values(?,?,?,?,?);
	`
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(job.TestType, job.Url, job.Interval, job.Timeout, job.CreatedAt)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func PutJob(job Job,  db *sql.DB)  error {
	q := `
		UPDATE jobs 
		SET test_type = ?,
			url = ?,
			interval = ?,
		    timeout = ?,
			created_at = ?
		WHERE job_id = ?
	`
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(job.TestType, job.Url, job.Interval, job.Timeout, job.CreatedAt, job.JobId)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func DeleteJob(id string,  db *sql.DB) error {
	q := `
		DELETE FROM jobs 
		WHERE job_id = ?
	`
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}