package dao

import (
	"github.com/jmoiron/sqlx"
	"pingr"
)


func GetRawTests(db *sqlx.DB) ([]pingr.GenericTest, error) {
	q := `
		SELECT * FROM tests
		ORDER BY test_name
	`
	var tests []pingr.GenericTest
	err := db.Select(&tests, q)
	return tests, err
}

func GetTests(db *sqlx.DB) ([]pingr.Test, error) {
	tests, err := GetRawTests(db)
	if err != nil {
		return nil, err
	}
	var parsedTests []pingr.Test
	for _, j := range tests {
		pTest, err := j.Impl()
		if err != nil {
			return nil, err
		}
		parsedTests = append(parsedTests, pTest)
	}

	return parsedTests, nil
}

func GetRawTest(id string, db *sqlx.DB) (test pingr.GenericTest, err error) {
	q := `
		SELECT * FROM tests 
		WHERE test_id = $1
	`
	err = db.Get(&test, q, id)
	if err != nil {
		return
	}

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
		INSERT INTO tests(test_id, test_name, test_type, url, interval, timeout, created_at, blob) 
		VALUES (:test_id,:test_name,:test_type,:url,:interval,:timeout,:created_at,:blob);
	`
	_, err := db.NamedExec(q, test)
	if err != nil {
		return err
	}

	return nil
}

func PutTest(test pingr.GenericTest,  db *sqlx.DB)  error {
	q := `
		UPDATE tests 
		SET test_name = :test_name,
		    test_type = :test_type,
			url = :url,
			interval = :interval,
		    timeout = :timeout,
			created_at = :created_at,
			blob = :blob
		WHERE test_id = :test_id
	`
	_, err := db.NamedExec(q, test)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTest(id string,  db *sqlx.DB) error {
	q := `
		DELETE FROM tests 
		WHERE test_id = $1
	`
	_, err := db.Exec(q, id)
	if err != nil {
		return err
	}

	return nil
}
/*
func (j Test) Parse() (parsedTest pingr.Test, err error) {
	switch j.TestType {
	case "SSH":
		var t pingr.SSHTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "TCP":
		var t pingr.TCPTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "TLS":
		var t pingr.TLSTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "Ping":
		var t pingr.PingTest
		t.BaseTest = j.BaseTest
		parsedTest = t
	case "HTTP":
		var t pingr.HTTPTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "DNS":
		var t pingr.DNSTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "Prometheus":
		var t pingr.PrometheusTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "PrometheusPush":
		var t pingr.PrometheusPushTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	case "HTTPPush":
		var t pingr.HTTPPushTest
		t.BaseTest = j.BaseTest
		err = json.Unmarshal(j.Blob, &t)
		if err != nil {
			return
		}
		parsedTest = t
	default:
		err = errors.New(j.TestType + " is not a valid test type")
	}
	return
}*/
