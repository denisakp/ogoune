package config

import (
	"log"
	"os"
)

type Config struct {
	RedisUrl    string
	DatabaseUrl string
}

func Load() Config {
	cfg := Config{
		RedisUrl:    GetEnv("REDIS_URL", "localhost:6379"),
		DatabaseUrl: GetEnv("DATABASE_URL", "postgres://denis:password@localhost:5432/pulse?sslmode=disable"),
	}
	return cfg
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func MustInit() Config {
	cfg := Load()

	dsn := cfg.DatabaseUrl
	if dsn == "" {
		log.Fatalf("DATABASE_URL environment variable is required")
	}

	return cfg
}
