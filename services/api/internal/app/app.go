package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RandomThacker/donna/services/api/internal/buildinfo"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/config"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/database"
	"github.com/RandomThacker/donna/services/api/internal/handler"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/router"
	"github.com/RandomThacker/donna/services/api/internal/server"
)

// Run wires dependencies, starts the HTTP server, and blocks until shutdown.
//
// Data flow: Handler → Business → Repository → Database
// Observability: all modules obtain loggers from the Logger Factory.
func Run(ctx context.Context, cfg *config.Config, logFactory *logger.Factory) error {
	appLog := logFactory.Module(constant.ModuleApp)
	httpLog := logFactory.Module(constant.ModuleHTTP)
	dbLog := logFactory.Module(constant.ModuleDatabase)

	pool, err := database.Connect(ctx, cfg)
	if err != nil {
		appLog.Error(ctx, "database connect failed", constant.LogAttrError, err)
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()
	dbLog.Info(ctx, "database pool ready")

	healthRepo := repository.NewHealthRepository(pool)
	healthSvc := business.NewHealthService(healthRepo, business.BuildInfo{
		Version:   buildinfo.Version,
		Commit:    buildinfo.Commit,
		BuildTime: buildinfo.BuildTime,
	})
	healthHandler := handler.NewHealthHandler(healthSvc)

	engine := router.New(router.Options{
		Environment:   cfg.App.Environment,
		CORSOrigins:   cfg.App.CORSOrigins,
		HTTPLogger:    httpLog,
		HealthHandler: healthHandler,
	})

	srv := server.New(cfg.App.Addr, engine, appLog)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		appLog.Info(ctx, "shutdown signal received", constant.LogAttrSignal, sig.String())
	case <-ctx.Done():
		appLog.Info(ctx, "shutdown context canceled", constant.LogAttrError, ctx.Err())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.App.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLog.Error(shutdownCtx, "shutdown failed", constant.LogAttrError, err)
		return fmt.Errorf("shutdown: %w", err)
	}

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	appLog.Info(context.Background(), "application stopped cleanly")
	return nil
}
