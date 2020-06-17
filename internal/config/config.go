package config

import (
	"github.com/caarlos0/env/v6"
	"log"
	"sync"
	"time"
)

type Config struct {
	Dev          bool          `env:"DEV" envDefault:"true"`
	Port         int           `env:"PORT" envDefault:"8080"`
	TermDuration time.Duration `env:"TERM_DURATION" envDefault:"20s"` // time allowed for graceful shutdown
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
