package config

import (
	"github.com/caarlos0/env/v6"
	"log"
	"sync"
	"time"
)

type Config struct {
	Dev           bool   `env:"DEV" envDefault:"false"`
	BaseUrl       string `env:"BASE_URL,required"`
	Port          int    `env:"PORT" envDefault:"8080"`
	SQLitePath    string `env:"SQLITE_PATH" envDefault:"pingr.sqlite"`
	SQLiteMigrate bool   `env:"SQLITE_MIGRATE" envDefault:"false"`

	BasicAuthUser string `env:"BASIC_AUTH_USER,required"`
	BasicAuthPass string `env:"BASIC_AUTH_PASS,required"`

	TermDuration time.Duration `env:"TERM_DURATION" envDefault:"20s"` // time allowed for graceful shutdown

	SMTPHost     string `env:"SMTP_HOST" envDefault:"smtp.gmail.com"`
	SMTPPort     int    `env:"SMTP_PORT" envDefault:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`

	AESKey string `env:"AES_KEY" envDefault:"6368616e676520746869732070617373776f726420746f206120736563726574"`

	MinDiscStorage uint64 `env:"MIN_DISC_STORAGE" envDefault:"5"` // GB:s
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
