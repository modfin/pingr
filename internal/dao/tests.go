package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)

type TestStatus struct {
	TestId       string `json:"test_id" db:"test_id"`
	TestName     string `json:"test_name" db:"test_name"`
	TestType     string `json:"test_type" db:"test_type"`
	Active       bool   `json:"active" db:"active"`
	Url          string `json:"url" db:"url"`
	StatusId     int    `json:"status_id" db:"status_id"`
	ResponseTime int    `json:"response_time" db:"response_time"`
}

type FullTestStatus struct {
	pingr.GenericTest
	StatusId int `json:"status_id" db:"status_id"`
}

func GetRawTests(db *sqlx.DB) ([]pingr.GenericTest, error) {
	q := `
		SELECT * FROM tests
		ORDER BY test_name
	`
	var tests []pingr.GenericTest
	err := db.Select(&tests, q)
	return tests, err
}

func GetRawTest(id string, db *sqlx.DB) (test pingr.GenericTest, err error) {
	q := `
		SELECT * FROM tests 
		WHERE test_id = $1
	`
	err = db.Get(&test, q, id)
	return
}

func GetTest(id string, db *sqlx.DB) (test pingr.Test, err error) {
	t, err := GetRawTest(id, db)
	if err != nil {
		return
	}
	test, err = t.Impl()
	return
}

func PostTest(test pingr.GenericTest, db *sqlx.DB) error {
	q := `
		INSERT INTO tests(test_id, test_name, test_type, url, interval, timeout, created_at, active, blob) 
		VALUES (:test_id,:test_name,:test_type,:url,:interval,:timeout,:created_at,:active,:blob);
	`
	_, err := db.NamedExec(q, test)
	return err
}

func PutTest(test pingr.GenericTest, db *sqlx.DB) error {
	q := `
		UPDATE tests 
		SET test_name = :test_name,
		    test_type = :test_type,
			url = :url,
			interval = :interval,
		    timeout = :timeout,
			created_at = :created_at,
		    active = :active,
			blob = :blob
		WHERE test_id = :test_id
	`
	_, err := db.NamedExec(q, test)
	return err
}

func DeleteTest(id string, db *sqlx.DB) error {
	q := `
		DELETE FROM tests 
		WHERE test_id = $1
	`
	_, err := db.Exec(q, id)
	return err
}

func GetTestStatus(id string, db *sqlx.DB) (FullTestStatus, error) {
	q := `
		SELECT t.*, l2.status_id
		FROM tests t
		INNER JOIN logs l2 ON l2.log_id IN (
		SELECT l1.log_id FROM logs l1 WHERE l1.test_id = t.test_id ORDER BY l1.created_at DESC LIMIT 1)
		WHERE t.test_id = $1
	`
	var testStatus FullTestStatus
	err := db.Get(&testStatus, q, id)
	return testStatus, err
}

func GetTestsStatus(db *sqlx.DB) ([]TestStatus, error) {
	q := `
		SELECT t.test_id, t.test_name, t.test_type, t.active, t.url, l2.status_id, l2.response_time
		FROM tests t
		INNER JOIN logs l2 ON l2.log_id IN (
		SELECT l1.log_id FROM logs l1 WHERE l1.test_id = t.test_id ORDER BY l1.created_at DESC LIMIT 1)
		ORDER BY t.test_name
	`
	var testStatus []TestStatus
	err := db.Select(&testStatus, q)
	if err != nil {
		return nil, err
	}
	return testStatus, nil

}

func DeactivateTest(testId string, db *sqlx.DB) error {
	q := `
		UPDATE tests
		SET active = 0
		WHERE test_id = $1
	`
	_, err := db.Exec(q, testId)

	return err
}

func ActivateTest(testId string, db *sqlx.DB) error {
	q := `
		UPDATE tests
		SET active = 1
		WHERE test_id = $1
	`
	_, err := db.Exec(q, testId)

	return err
}
