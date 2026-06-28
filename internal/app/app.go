package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/adesubomi/pigeon-server/config"
	"github.com/adesubomi/pigeon-server/infra/db"
	otelInfra "github.com/adesubomi/pigeon-server/infra/otel"
	redisInfra "github.com/adesubomi/pigeon-server/infra/redis"
	"github.com/adesubomi/pigeon-server/internal/app/handlers"
	"github.com/adesubomi/pigeon-server/internal/app/repo"
	"github.com/adesubomi/pigeon-server/internal/app/services"
	"github.com/adesubomi/pigeon-server/internal/domain/push"
)

type App struct {
	cfg          *config.Config
	logger       *slog.Logger
	db           *db.DB
	redis        *redisInfra.Client
	otelShutdown otelInfra.ShutdownFunc
	router       http.Handler
	server       *http.Server
	errCh        chan error
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*App, error) {
	otelShutdown, err := otelInfra.Setup(ctx, cfg)
	if err != nil {
		return nil, err
	}

	store, err := db.Connect(ctx, cfg)
	if err != nil {
		_ = otelShutdown(ctx)
		return nil, err
	}

	redisClient, err := redisInfra.Connect(ctx, cfg)
	if err != nil {
		_ = store.Close()
		_ = otelShutdown(ctx)
		return nil, err
	}

	pushHub := push.NewHub()
	authRepo := repo.NewAuthRepo(store.Gorm)
	deviceRepo := repo.NewDeviceRepo(store.Gorm)
	endpointRepo := repo.NewEndpointRepo(store.Gorm)
	eventRepo := repo.NewEventRepo(store.Gorm)
	pushRepo := repo.NewPushRepo(store.Gorm)

	pushSvc := services.NewPushSvc(pushRepo, pushHub)
	authSvc := services.NewAuth(authRepo, cfg)
	endpointSvc := services.NewEndpointSvc(endpointRepo)
	deviceSvc := services.NewDevice(deviceRepo)
	eventSvc := services.NewEvent(eventRepo, pushSvc)

	authHandler := handlers.NewHandler(authSvc, deviceSvc, endpointSvc, eventSvc, pushSvc)

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
