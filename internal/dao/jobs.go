package dao

import (
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"pingr"
)


type Job struct {
	pingr.BaseJob
	Blob []byte `json:"blob" db:"blob"`
}

func (j Job) Parse() (parsedJob pingr.Job, err error) {
	switch j.TestType {
	case "SSH":
		var t pingr.SSHTest
		t.BaseJob = j.BaseJob
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedJob = t
	case "TCP":
		var t pingr.TCPTest
		t.BaseJob = j.BaseJob
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedJob = t
	case "TLS":
		var t pingr.TLSTest
		t.BaseJob = j.BaseJob
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedJob = t
	case "Ping":
		var t pingr.PingTest
		t.BaseJob = j.BaseJob
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedJob = t
	case "HTTP":
		var t pingr.HTTPTest
		t.BaseJob = j.BaseJob
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedJob = t
	default:
		err = errors.New(j.TestType + " is not a vaild test type")
	}
	return
}

func GetJobs(db *sqlx.DB) ([]pingr.Job, error) {
	q := `
		SELECT * FROM jobs
	`
	var jobs []Job
	err := db.Select(&jobs, q)
	if err != nil {
		return nil, err
	}

	var parsedJobs []pingr.Job
	for _, j := range jobs {
		pJob, err := j.Parse()
		if err != nil {
			return nil, err
		}
		parsedJobs = append(parsedJobs, pJob)
	}

	return parsedJobs, nil
}

func GetJob(id uint64, db *sqlx.DB) (_job pingr.Job, err error) {
	q := `
		SELECT * FROM jobs WHERE job_id = $1
	`
	var j Job
	err = db.Get(&j, q, id)
	if err != nil {
		return
	}

	_job, err = j.Parse()
	return
}

func GetJobLogs(id uint64, db *sqlx.DB) ([]pingr.Log, error) {
	q := `
		SELECT * FROM logs WHERE job_id = $1
	`
	var logs []pingr.Log

	err := db.Select(&logs, q, id)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func PostJob(job Job, db *sqlx.DB) error {
	q := `
		INSERT INTO jobs(test_type, url, interval, timeout, created_at, blob) 
		VALUES (:test_type,:url,:interval,:timeout,:created_at,:blob);
	`
	_, err := db.NamedExec(q, job)
	if err != nil {
		return err
	}

	return nil
}

func PutJob(job Job,  db *sqlx.DB)  error {
	q := `
		UPDATE jobs 
		SET test_type = :test_type,
			url = :url,
			interval = :interval,
		    timeout = :timeout,
			created_at = :created_at,
			blob = :blob
		WHERE job_id = :job_id
	`
	_, err := db.NamedExec(q, job)
	if err != nil {
		return err
	}

	return nil
}

func DeleteJob(id uint64,  db *sqlx.DB) error {
	q := `
		DELETE FROM jobs 
		WHERE job_id = $1
	`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}