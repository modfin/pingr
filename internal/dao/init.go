package dao

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"os"
	"pingr/internal/config"
)

func InitDB() (*sqlx.DB, error) {
	dbfile := config.Get().SQLitePath

	var isnew bool
	if !fileExists(dbfile) {
		isnew = true
		file, err := os.Create(dbfile)
		if err != nil {
			return nil, err
		}
		file.Close()
	}

	db, err := sqlx.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}

	if isnew || config.Get().SQLiteMigrate {
		err = migrateSchema(db)
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrateSchema(db *sqlx.DB) error {
	log.Info("Migrating sql schema")
	_, err := db.Exec(_schema_v0_up)
	if err != nil {
		return err
	}
	var version int
	err = db.Get(&version, "SELECT max(version) FROM _schema")
	if err != nil {
		return err
	}
	log.Info("  - Currently at ", version)

	switch version {
	case 0:
		log.Info("  - Migrating to ", 1)
		_, err := db.Exec(_schema_v1_up)
		if err != nil {
			return err
		}
	}

	return nil
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
