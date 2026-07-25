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
	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
	"github.com/RandomThacker/donna/services/api/internal/handler"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/router"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/RandomThacker/donna/services/api/internal/server"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/jackc/pgx/v5/pgxpool"
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

	identityLog := logFactory.Module(constant.ModuleIdentity)
	userRepo := repository.NewUserRepository(pool)
	userSvc := business.NewUserService(userRepo, identityLog)
	userHandler := handler.NewUserHandler(userSvc, identityLog)

	authLog := logFactory.Module(constant.ModuleAuth)
	authParts, err := wireAuth(cfg, pool, userSvc, userRepo, authLog)
	if err != nil {
		appLog.Error(ctx, "auth wiring failed", constant.LogAttrError, err)
		return fmt.Errorf("auth: %w", err)
	}

	engine := router.New(router.Options{
		Environment:   cfg.App.Environment,
		CORSOrigins:   cfg.App.CORSOrigins,
		HTTPLogger:    httpLog,
		HealthHandler: healthHandler,
		UserHandler:   userHandler,
		AuthHandler:   authParts.handler,
		MeHandler:     handler.NewMeHandler(userSvc),
		TokenIssuer:   authParts.issuer,
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

type authWire struct {
	handler *handler.AuthHandler
	issuer  *session.Issuer
}

func wireAuth(
	cfg *config.Config,
	pool *pgxpool.Pool,
	userSvc *business.UserService,
	userRepo repository.UserRepository,
	authLog *logger.Logger,
) (authWire, error) {
	sealKey, err := seal.KeyFromSecret(cfg.App.CredentialsKey)
	if err != nil {
		return authWire{}, fmt.Errorf("credentials key: %w", err)
	}
	tokenIssuer, err := session.NewIssuer(cfg.App.JWTSecret, cfg.App.JWTExpiry)
	if err != nil {
		return authWire{}, fmt.Errorf("jwt issuer: %w", err)
	}

	var googleClient business.GoogleOAuthClient
	if cfg.API.GoogleOAuth.ClientID != "" && cfg.API.GoogleOAuth.ClientSecret != "" {
		client, err := googleoauth.NewClient(googleoauth.Config{
			ClientID:     cfg.API.GoogleOAuth.ClientID,
			ClientSecret: cfg.API.GoogleOAuth.ClientSecret,
			RedirectURL:  cfg.API.GoogleOAuth.RedirectURL,
			AuthURL:      cfg.API.GoogleOAuth.AuthURL,
			TokenURL:     cfg.API.GoogleOAuth.TokenURL,
			UserInfoURL:  cfg.API.GoogleOAuth.UserInfoURL,
			Scopes:       cfg.API.GoogleOAuth.Scopes,
			Timeout:      cfg.API.GoogleOAuth.Timeout,
		})
		if err != nil {
			return authWire{}, fmt.Errorf("google oauth client: %w", err)
		}
		googleClient = client
		authLog.Info(context.Background(), "google oauth client configured")
	} else {
		authLog.Warn(context.Background(), "google oauth not configured; /auth/google will reject requests")
	}

	authSvc := business.NewAuthService(business.AuthServiceDeps{
		Users:      userSvc,
		UserRepo:   userRepo,
		Identities: repository.NewAuthIdentityRepository(pool),
		Accounts:   repository.NewConnectedAccountRepository(pool),
		Secrets:    repository.NewCredentialSecretRepository(pool),
		Tx:         repository.NewTxManager(pool),
		Google:     googleClient,
		State:      oauthstate.NewManager(cfg.App.JWTSecret),
		Tokens:     tokenIssuer,
		SealKey:    sealKey,
		Log:        authLog,
	})

	return authWire{
		handler: handler.NewAuthHandler(authSvc, authLog, handler.AuthHandlerConfig{
			FrontendSuccessURL: cfg.App.FrontendSuccessURL,
			CookieSecure:       cfg.App.CookieSecure,
			CookieMaxAge:       cfg.App.JWTExpiry,
			Tokens:             tokenIssuer,
		}),
		issuer: tokenIssuer,
	}, nil
}
