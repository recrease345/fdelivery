package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBMaxConns        int           `env:"DB_MAX_CONNS" env-required:"true"`
	DBMinConns        int           `env:"DB_MIN_CONNS" env-required:"true"`
	DBMaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME" env-required:"true"`
	DBMaxConnIdleTime time.Duration `env:"DB_MAX_CONN_IDLE_TIME" env-required:"true"`
	DBURL             string        `env:"DB_URL" env-required:"true"`
}

func LoadCFG() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
