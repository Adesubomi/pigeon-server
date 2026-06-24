package app

import (
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"time"

	"github.com/adesubomi/pigeon-server/internal/app/handlers"
	"github.com/adesubomi/pigeon-server/internal/app/services"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type routesConfig struct {
	authSvc     *services.AuthService
	deviceSvc   *services.DeviceService
	eventSvc    *services.EventService
	endpointSvc *services.EndpointService
	handlers    *handlers.Handler
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
	r.Post("/hooks/{slug}", cfg.handlers.Events().ReceiveWebhook)

	r.Get("/auth/github", cfg.handlers.Auth().GitHubLogin)
	r.Post("/auth/github/exchange", cfg.handlers.Auth().GitHubExchange)
	r.Post("/auth/logout", cfg.handlers.Auth().Logout)

	r.With(cfg.authSvc.RequireUser).Get("/me", cfg.handlers.Auth().Me)

	r.Group(func(r chi.Router) {
		r.Use(cfg.authSvc.RequireUser)
		r.Post("/endpoints", cfg.handlers.Endpoints().CreateEndpoint)
		r.Get("/endpoints", cfg.handlers.Endpoints().ListEndpoints)
		r.Get("/endpoints/{id}", cfg.handlers.Endpoints().GetEndpoint)
		r.Patch("/endpoints/{id}", cfg.handlers.Endpoints().UpdateEndpoint)
		r.Delete("/endpoints/{id}", cfg.handlers.Endpoints().DeleteEndpoint)
		r.Post("/endpoints/{id}/pairing-codes", cfg.handlers.Endpoints().GeneratePairingCode)
		r.Get("/endpoints/{id}/devices", cfg.handlers.Endpoints().ListEndpointDevices)
		r.Get("/endpoints/{id}/events", cfg.handlers.Endpoints().ListEndpointEvents)
		r.Get("/events/{id}", cfg.handlers.Events().GetEvent)
		r.Post("/events/{id}/replay", cfg.handlers.Events().ReplayEvent)
	})

	r.Post("/devices/pair", cfg.handlers.Devices().PairDevice)
	r.Group(func(r chi.Router) {
		r.Use(cfg.deviceSvc.RequireDevice)
		r.Post("/devices/heartbeat", cfg.handlers.Devices().Heartbeat)
		r.Patch("/devices/{id}", cfg.handlers.Devices().UpdateDevice)
		r.Delete("/devices/{id}", cfg.handlers.Devices().DeleteDevice)
		r.Get("/stream", cfg.handlers.Devices().Stream)
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
