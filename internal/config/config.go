package config

import (
	"github.com/caarlos0/env/v6"
	"log"
	"sync"
	"time"
)

type Config struct {
	Dev           bool   `env:"DEV" envDefault:"false"`
	Port          int    `env:"PORT" envDefault:"8080"`
	SQLitePath    string `env:"SQLITE_PATH" envDefault:"pingr.sqlite"`
	SQLiteMigrate bool   `env:"SQLITE_MIGRATE" envDefault:"false"`

	TermDuration time.Duration `env:"TERM_DURATION" envDefault:"20s"` // time allowed for graceful shutdown

	SMTPHost     string `env:"SMTP_HOST" envDefault:"smtp.gmail.com"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`

	MinDiscStorage uint64 `env:"MIN_DISC_STORAGE" envDefault:"5"`
}

var (
	once sync.Once
	cfg  Config
)

func Get() Config {
	once.Do(func() {
		if err := env.Parse(&cfg); err != nil {
			log.Panic("Couldn't parse AppConfig from env: ", err)
		}
	})
	return cfg
}

func Reload() error {
	c := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}
	cfg = c
	return nil
}
