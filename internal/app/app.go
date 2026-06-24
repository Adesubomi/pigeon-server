package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/adesubomi/pigeon-server/config"
	"github.com/adesubomi/pigeon-server/infra/db"
	infraotel "github.com/adesubomi/pigeon-server/infra/otel"
	infraredis "github.com/adesubomi/pigeon-server/infra/redis"
	"github.com/adesubomi/pigeon-server/internal/app/handlers"
	"github.com/adesubomi/pigeon-server/internal/app/services"
	"github.com/adesubomi/pigeon-server/internal/domain/push"
)

type App struct {
	cfg          *config.Config
	logger       *slog.Logger
	db           *db.DB
	redis        *infraredis.Client
	otelShutdown infraotel.ShutdownFunc
	router       http.Handler
	server       *http.Server
	errCh        chan error
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	otelShutdown, err := infraotel.Setup(ctx, cfg)
	if err != nil {
		return nil, err
	}

	store, err := db.Connect(ctx, cfg)
	if err != nil {
		_ = otelShutdown(ctx)
		return nil, err
	}

	redisClient, err := infraredis.Connect(ctx, cfg)
	if err != nil {
		_ = store.Close()
		_ = otelShutdown(ctx)
		return nil, err
	}

	pushHub := push.NewHub()
	pushSvc := services.NewPushSvc(store.Gorm, pushHub)
	authSvc := services.NewAuth(store.Gorm, cfg)
	endpointSvc := services.NewEndpointSvc(store.Gorm)
	deviceSvc := services.NewDevice(store.Gorm)
	eventSvc := services.NewEvent(store.Gorm, pushSvc)

	authHandler := handlers.NewHandler(authSvc)

	app := &App{
		cfg:          cfg,
		logger:       logger,
		db:           store,
		redis:        redisClient,
		otelShutdown: otelShutdown,
		errCh:        make(chan error, 1),
	}

	app.router = app.routes(routesConfig{
		authSvc:     authSvc,
		deviceSvc:   deviceSvc,
		endpointSvc: endpointSvc,
		eventSvc:    eventSvc,

		handlers: authHandler,
	})
	app.server = &http.Server{
		Addr:    cfg.HTTPAddr(),
		Handler: app.router,
	}

	return app, nil
}

func (a *App) Errors() <-chan error {
	return a.errCh
}
