package main

import (
	"pingr/internal/config"
	"pingr/internal/logging"
	"pingr/internal/resources"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	resources.Init(closing)

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
