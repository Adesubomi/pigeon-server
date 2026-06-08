package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDotEnv(t *testing.T) {
	cleanConfigEnv(t)
	t.Chdir(t.TempDir())
	writeDotEnv(t, "APP_NAME=from-dotenv\nAPP_PORT=19090\n")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppName != "from-dotenv" {
		t.Errorf("AppName = %q, want %q", cfg.AppName, "from-dotenv")
	}
	if cfg.AppPort != 19090 {
		t.Errorf("AppPort = %d, want %d", cfg.AppPort, 19090)
	}
}

func TestLoadDoesNotOverrideProcessEnvironment(t *testing.T) {
	cleanConfigEnv(t)
	t.Chdir(t.TempDir())
	writeDotEnv(t, "APP_NAME=from-dotenv\n")
	t.Setenv("APP_NAME", "from-process")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppName != "from-process" {
		t.Errorf("AppName = %q, want %q", cfg.AppName, "from-process")
	}
}

func writeDotEnv(t *testing.T, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(".", ".env"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
}

func cleanConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"APP_NAME",
		"APP_ENV",
		"APP_PORT",
		"SERVICE_NAME",
		"APP_KEY",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"DB_SSLMODE",
		"REDIS_HOST",
		"REDIS_PORT",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"GITHUB_CLIENT_ID",
		"GITHUB_CLIENT_SECRET",
		"WEB_APP_URL",
		"OAUTH_REDIRECT_ALLOWLIST",
		"AUTH_ACCESS_TOKEN_TTL",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
	}
	for _, key := range keys {
		value, exists := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if exists {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		})
	}
}
