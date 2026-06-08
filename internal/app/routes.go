package app

import (
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/internal/domain/device"
	"github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/internal/domain/push"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type routesConfig struct {
	authService     *auth.Service
	deviceService   *device.Service
	authHandler     *auth.Handler
	endpointHandler *endpoint.Handler
	deviceHandler   *device.Handler
	eventHandler    *event.Handler
	pushHandler     *push.Handler
}

func (a *App) routes(cfg routesConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(a.corsMiddleware)
	r.Use(a.recoverMiddleware)
	r.Use(a.loggingMiddleware)
	r.Use(func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http.request")
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		respond.OK(w, map[string]string{"status": "ok"})
	})
	r.Post("/hooks/{slug}", cfg.eventHandler.ReceiveWebhook)

	r.Get("/auth/github", cfg.authHandler.GitHubLogin)
	r.Post("/auth/github/exchange", cfg.authHandler.GitHubExchange)
	r.Post("/auth/logout", cfg.authHandler.Logout)
	r.With(cfg.authService.RequireUser).Get("/me", cfg.authHandler.Me)

	r.Group(func(r chi.Router) {
		r.Use(cfg.authService.RequireUser)
		r.Post("/endpoints", cfg.endpointHandler.CreateEndpoint)
		r.Get("/endpoints", cfg.endpointHandler.ListEndpoints)
		r.Get("/endpoints/{id}", cfg.endpointHandler.GetEndpoint)
		r.Patch("/endpoints/{id}", cfg.endpointHandler.UpdateEndpoint)
		r.Delete("/endpoints/{id}", cfg.endpointHandler.DeleteEndpoint)
		r.Post("/endpoints/{id}/pairing-codes", cfg.endpointHandler.GeneratePairingCode)
		r.Get("/endpoints/{id}/devices", cfg.endpointHandler.ListEndpointDevices)
		r.Get("/endpoints/{id}/events", cfg.endpointHandler.ListEndpointEvents)
		r.Get("/events/{id}", cfg.eventHandler.GetEvent)
		r.Post("/events/{id}/replay", cfg.eventHandler.ReplayEvent)
	})

	r.Post("/devices/pair", cfg.deviceHandler.PairDevice)
	r.Group(func(r chi.Router) {
		r.Use(cfg.deviceService.RequireDevice)
		r.Post("/devices/heartbeat", cfg.deviceHandler.Heartbeat)
		r.Patch("/devices/{id}", cfg.deviceHandler.UpdateDevice)
		r.Delete("/devices/{id}", cfg.deviceHandler.DeleteDevice)
		r.Get("/stream", cfg.pushHandler.Stream)
	})

	return r
}

func (a *App) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == webAppOrigin(a.cfg.WebAppURL) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Pigeon-Device-Token")
		}
		if r.Method == http.MethodOptions {
			respond.NoContent(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func webAppOrigin(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func (a *App) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				a.logger.Error("panic recovered", "panic", value, "stack", string(debug.Stack()))
				respond.Error(w, apperr.Internal(http.ErrAbortHandler))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *App) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		a.logger.Info("http request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", recorder.status),
			slog.Duration("duration", time.Since(start)),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
