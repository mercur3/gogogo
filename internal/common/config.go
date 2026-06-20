package common

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	MaxBodySize int64

	// DB
	DbUser     string
	DbPassword string
	DbHost     string
	DbPort     string
	DbName     string
}

const (
	maxBodySizeEnv = "MAX_BODY_SIZE"
	dbUserEnv      = "DB_USER"
	dbPasswordEnv  = "DB_PASSWORD"
	dbHostEnv      = "DB_HOST"
	dbPortEnv      = "DB_PORT"
	dbNameEnv      = "DB_NAME"
)

func ParseConfigs() (Config, error) {
	cfg := Config{}

	maxBodySize, ok := os.LookupEnv(maxBodySizeEnv)
	if !ok {
		logMissing(maxBodySizeEnv)
		cfg.MaxBodySize = 1 << 20
	} else {
		val, err := strconv.ParseInt(maxBodySize, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse value=%s: %w", maxBodySize, err)
		}
		cfg.MaxBodySize = int64(val)
	}

	err := requiredEnv(dbUserEnv, &cfg, func(cfg *Config, env string) { cfg.DbUser = env })
	if err != nil {
		return cfg, err
	}

	err = requiredEnv(dbPasswordEnv, &cfg, func(cfg *Config, env string) { cfg.DbPassword = env })
	if err != nil {
		return cfg, err
	}

	err = requiredEnv(dbHostEnv, &cfg, func(cfg *Config, env string) { cfg.DbHost = env })
	if err != nil {
		return cfg, err
	}

	err = requiredEnv(dbPortEnv, &cfg, func(cfg *Config, env string) { cfg.DbPort = env })
	if err != nil {
		return cfg, err
	}

	err = requiredEnv(dbNameEnv, &cfg, func(cfg *Config, env string) { cfg.DbName = env })
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func logMissing(env string) {
	slog.Debug("Env not set", slog.String("env", env))
}

func requiredEnv(env string, cfg *Config, set func(cfg *Config, env string)) error {
	envValue, ok := os.LookupEnv(env)
	if !ok {
		logMissing(env)
		return fmt.Errorf("missing required env: %s", env)
	}

	set(cfg, envValue)
	return nil
}
