package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adesubomi/pigeon-server/config"
)

func TestCORSMiddlewareAllowsConfiguredWebOrigin(t *testing.T) {
	app := &App{
		cfg:    &config.Config{WebAppURL: "http://localhost:3000"},
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	handler := app.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/me", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Headers"); got != "Authorization,Content-Type,X-Pigeon-Device-Token" {
		t.Fatalf("allow headers = %q", got)
	}
}

func TestCORSMiddlewareDoesNotAllowUntrustedOrigin(t *testing.T) {
	app := &App{cfg: &config.Config{WebAppURL: "http://localhost:3000"}}
	handler := app.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Origin", "https://attacker.example")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q", got)
	}
}
