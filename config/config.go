package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName                  string
	AppEnv                   string
	AppPort                  int
	ServiceName              string
	AppKey                   string
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
	WebAppURL                string
	OAuthRedirectAllowlist   []string
	AuthAccessTokenTTL       time.Duration
	OTELExporterOTLPEndpoint string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

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
	webAppURL := strings.TrimRight(stringFromEnv("WEB_APP_URL", "http://localhost:3000"), "/")
	if _, err := parseWebAppURL(webAppURL); err != nil {
		return nil, err
	}

	return &Config{
		AppName:                  stringFromEnv("APP_NAME", "pigeon"),
		AppEnv:                   stringFromEnv("APP_ENV", "development"),
		AppPort:                  appPort,
		ServiceName:              stringFromEnv("SERVICE_NAME", "pigeon-server"),
		AppKey:                   stringFromEnv("APP_KEY", "development-app-key-change-me"),
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
		WebAppURL:                webAppURL,
		OAuthRedirectAllowlist:   csvFromEnv("OAUTH_REDIRECT_ALLOWLIST"),
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

func csvFromEnv(key string) []string {
	var values []string
	for value := range strings.SplitSeq(os.Getenv(key), ",") {
		if value = strings.TrimSpace(value); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func parseWebAppURL(value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("WEB_APP_URL must be an absolute HTTP(S) origin")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("WEB_APP_URL must use HTTP or HTTPS")
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return nil, fmt.Errorf("WEB_APP_URL must not include a path")
	}
	return parsed, nil
}
