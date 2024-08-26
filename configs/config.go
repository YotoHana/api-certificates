package config

import (
	"os"
)

type Config struct {
	DbConn string
	SecretKey string
}

func New() *Config {
	return &Config{
		DbConn: getEnv("DBCONN", "postgres://adminfront:123@localhost:5432/adminfront"),
		SecretKey: getEnv("SECRET_KEY", "secret_key"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists{
		return value
	}
	return defaultVal
}