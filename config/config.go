package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv                   string
	AppPort                  int
	AppKey                   string
	PublicBaseURL            string
	DBHost                   string
	DBPort                   int
	DBUser                   string
	DBPassword               string
	DBName                   string
	DBSSLMode                string
	RedisHost                string
	RedisPort                int
	RedisPassword            string
	RedisDB                  int
	GitHubClientID           string
	GitHubClientSecret       string
	AuthAccessTokenTTL       time.Duration
	OTELExporterOTLPEndpoint string
}

func Load() (*Config, error) {
	appPort, err := intFromEnv("APP_PORT", 8080)
	if err != nil {
		return nil, err
	}
	dbPort, err := intFromEnv("DB_PORT", 5432)
	if err != nil {
		return nil, err
	}
	redisPort, err := intFromEnv("REDIS_PORT", 6379)
	if err != nil {
		return nil, err
	}
	redisDB, err := intFromEnv("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}
	authAccessTokenTTL, err := durationFromEnv("AUTH_ACCESS_TOKEN_TTL", time.Hour)
	if err != nil {
		return nil, err
	}

	return &Config{
		AppEnv:                   stringFromEnv("APP_ENV", "development"),
		AppPort:                  appPort,
		AppKey:                   stringFromEnv("APP_KEY", "development-app-key-change-me"),
		PublicBaseURL:            stringFromEnv("PUBLIC_BASE_URL", "http://localhost:8080"),
		DBHost:                   stringFromEnv("DB_HOST", "localhost"),
		DBPort:                   dbPort,
		DBUser:                   stringFromEnv("DB_USER", "postgres"),
		DBPassword:               stringFromEnv("DB_PASSWORD", "postgres"),
		DBName:                   stringFromEnv("DB_NAME", "pigeon"),
		DBSSLMode:                stringFromEnv("DB_SSLMODE", "disable"),
		RedisHost:                stringFromEnv("REDIS_HOST", "localhost"),
		RedisPort:                redisPort,
		RedisPassword:            os.Getenv("REDIS_PASSWORD"),
		RedisDB:                  redisDB,
		GitHubClientID:           os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:       os.Getenv("GITHUB_CLIENT_SECRET"),
		AuthAccessTokenTTL:       authAccessTokenTTL,
		OTELExporterOTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}, nil
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

func (c *Config) RedisAddr() string {
	return net.JoinHostPort(c.RedisHost, strconv.Itoa(c.RedisPort))
}

func stringFromEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intFromEnv(key string, fallback int) (int, error) {
	if value := os.Getenv(key); value != "" {
		return strconv.Atoi(value)
	}
	return fallback, nil
}

func durationFromEnv(key string, fallback time.Duration) (time.Duration, error) {
	if value := os.Getenv(key); value != "" {
		return time.ParseDuration(value)
	}
	return fallback, nil
}
