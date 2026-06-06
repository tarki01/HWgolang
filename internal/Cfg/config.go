package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	SMTPHost   string
	SMTPPort   int
	SMTPUser   string
	SMTPPass   string
	PGPKey     string
	HMACSecret string
	ServerPort string
	LogLevel   string
}

func Load() *Config {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "13d3791semenek"),
		DBName:     getEnv("DB_NAME", "bankdb"),
		JWTSecret:  getEnv("JWT_SECRET", "ssseeeekret"),
		SMTPHost:   getEnv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:   smtpPort,
		SMTPUser:   getEnv("SMTP_USER", ""),
		SMTPPass:   getEnv("SMTP_PASS", ""),
		PGPKey:     getEnv("PGP_KEY", "default-pgp-passphrase"),
		HMACSecret: getEnv("HMAC_SECRET", "default-hmac-secret"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
