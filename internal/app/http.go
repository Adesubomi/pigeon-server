package app

import (
	"context"
	"errors"
	"net/http"
)

func (a *App) Start() error {
	go func() {
		a.logger.Info("starting HTTP server", "addr", a.cfg.HTTPAddr())
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.errCh <- err
		}
	}()
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down HTTP server")
	return errors.Join(
		a.server.Shutdown(ctx),
		a.redis.Close(),
		a.db.Close(),
		a.otelShutdown(ctx),
	)
}
