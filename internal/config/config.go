package config

import (
	"net"
	"net/url"
	"os"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	CookieSecure bool
	ServerPort   string
}

func Load() Config {
	return Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "notes"),
		DBPassword:   getEnv("DB_PASSWORD", "notes"),
		DBName:       getEnv("DB_NAME", "notes_app"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		CookieSecure: getEnv("COOKIE_SECURE", "false") == "true",
		ServerPort:   getEnv("SERVER_PORT", "8080"),
	}
}

func (c Config) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   net.JoinHostPort(c.DBHost, c.DBPort),
		Path:   "/" + c.DBName,
	}
	q := u.Query()
	q.Set("sslmode", c.DBSSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
