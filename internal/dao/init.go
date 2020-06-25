package dao

import (
	"database/sql"
	"github.com/gchaincl/dotsql"
	"os"
)

func InitDB() (*sql.DB, error) {
	if !fileExists("data.db") {
		file, err := os.Create("data.db")
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}

	err = setupTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupTables(db *sql.DB) error {
	dot, err := dotsql.LoadFromFile("./_schema/001.tables.up.sql")
	if err != nil {
		return err
	}

	_, err = dot.Exec(db, "create-jobs-table")
	if err != nil {
		return err
	}

	_, err = dot.Exec(db, "create-logs-table")
	if err != nil {
		return err
	}

	return nil
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}