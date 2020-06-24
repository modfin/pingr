package main

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"pingr/internal/config"
	"pingr/internal/logging"
	"pingr/internal/resources"
	"syscall"
	"time"
	"github.com/gchaincl/dotsql"
)

func main() {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())
	closing := ctx.Done()
	defer cancel()

	go signaling(cancel)
	logging.SetDefault()

	log.WithField("pid", os.Getpid()).Info("Starting pingr")

	db, err := initDB()
	if err != nil {
		log.Fatal("Could not load the database: ", err)
	}
	defer db.Close()

	log.WithField("pid", os.Getpid()).Info("DB initialized")

	resources.Init(closing, db)

	log.Info("Terminating service")
}

func signaling(cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var exit, exitCancel = context.WithCancel(context.Background())
	for {
		select {
		case <-exit.Done():
			log.Warn("Forcing termination")
			os.Exit(1)
		case sig := <-signals:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info("Got SIGINT/SIGTERM, exiting..")
				cancel()
				cancel = exitCancel
				time.AfterFunc(config.Get().TermDuration, cancel)

			case syscall.SIGHUP:
				log.Info("Got SIGHUP, reloading.")
				err := config.Reload()
				if err != nil {
					log.WithError(err).Warn("could not reload config")
				}
			}
		}
	}
}

func initDB() (*sql.DB, error) {
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
