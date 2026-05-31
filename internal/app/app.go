package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/adesubomi/pigeon-server/config"
	"github.com/adesubomi/pigeon-server/infra/db"
	infraotel "github.com/adesubomi/pigeon-server/infra/otel"
	infraredis "github.com/adesubomi/pigeon-server/infra/redis"
	"github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/internal/domain/device"
	"github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/internal/domain/event"
	"github.com/adesubomi/pigeon-server/internal/domain/push"
	"github.com/adesubomi/pigeon-server/pkg/clock"
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

	realClock := clock.RealClock{}
	pushHub := push.NewHub()
	pushService := push.NewService(store.Gorm, pushHub)
	authService := auth.NewService(store.Gorm, cfg)
	endpointService := endpoint.NewService(store.Gorm, realClock)
	deviceService := device.NewService(store.Gorm, realClock)
	eventService := event.NewService(store.Gorm, realClock, pushService)

	authHandler := auth.NewHandler(authService)
	endpointHandler := endpoint.NewHandler(endpointService)
	deviceHandler := device.NewHandler(deviceService)
	eventHandler := event.NewHandler(eventService)
	pushHandler := push.NewHandler(pushService)

	app := &App{
		cfg:          cfg,
		logger:       logger,
		db:           store,
		redis:        redisClient,
		otelShutdown: otelShutdown,
		errCh:        make(chan error, 1),
	}

	app.router = app.routes(routesConfig{
		authService:     authService,
		deviceService:   deviceService,
		authHandler:     authHandler,
		endpointHandler: endpointHandler,
		deviceHandler:   deviceHandler,
		eventHandler:    eventHandler,
		pushHandler:     pushHandler,
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
