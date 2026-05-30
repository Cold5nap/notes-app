package config

import (
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	SessionSecret string
	ServerPort    string
}

func Load() Config {
	return Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "notes"),
		DBPassword:    getEnv("DB_PASSWORD", "notes"),
		DBName:        getEnv("DB_NAME", "notes_app"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-secret"),
		ServerPort:    getEnv("SERVER_PORT", "8080"),
	}
}

func (c Config) DSN() string {
	return "postgres://" + c.DBUser + ":" + c.DBPassword +
		"@" + c.DBHost + ":" + c.DBPort +
		"/" + c.DBName + "?sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
