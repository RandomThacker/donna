package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/buildinfo"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/calendarprovider"
	"github.com/RandomThacker/donna/services/api/internal/chat"
	"github.com/RandomThacker/donna/services/api/internal/config"
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/database"
	"github.com/RandomThacker/donna/services/api/internal/googlecalendar"
	"github.com/RandomThacker/donna/services/api/internal/googleoauth"
	"github.com/RandomThacker/donna/services/api/internal/handler"
	"github.com/RandomThacker/donna/services/api/internal/icscalendar"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/microsoftcalendar"
	"github.com/RandomThacker/donna/services/api/internal/microsoftoauth"
	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
	"github.com/RandomThacker/donna/services/api/internal/repository"
	"github.com/RandomThacker/donna/services/api/internal/router"
	"github.com/RandomThacker/donna/services/api/internal/scheduler"
	"github.com/RandomThacker/donna/services/api/internal/scheduler/calendarsync"
	"github.com/RandomThacker/donna/services/api/internal/seal"
	"github.com/RandomThacker/donna/services/api/internal/server"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/google/uuid"
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

	calendarLog := logFactory.Module(constant.ModuleCalendar)
	calendarParts, err := wireCalendar(cfg, pool, authParts, calendarLog)
	if err != nil {
		appLog.Error(ctx, "calendar wiring failed", constant.LogAttrError, err)
		return fmt.Errorf("calendar: %w", err)
	}
	if calendarParts.service != nil && calendarParts.integrations != nil {
		calendarParts.integrations.SetOnAccountReady(calendarParts.service.BootstrapCalendarSyncJob)
		calendarParts.integrations.SetSyncAccount(func(ctx context.Context, accountID uuid.UUID) (business.CalendarPipelineResult, error) {
			return calendarParts.service.SyncPipelineForAccount(ctx, accountID, constant.CalendarSyncTriggerManual)
		})
		authParts.service.SetLoginCalendarLinker(calendarParts.integrations)
	}

	taskLog := logFactory.Module(constant.ModuleTask)
	taskRepo := repository.NewTaskRepository(pool)
	occurrenceRepo := repository.NewTaskOccurrenceRepository(pool)
	dailyNoteRepo := repository.NewDailyNoteRepository(pool)
	taskTagRepo := repository.NewTaskTagRepository(pool)
	taskSvc := business.NewTaskJournalService(taskRepo, occurrenceRepo, dailyNoteRepo, taskTagRepo, userRepo)

	donnaEventRepo := repository.NewDonnaEventRepository(pool)
	donnaReminderRepo := repository.NewDonnaReminderRepository(pool)
	calendarEventsRepo := repository.NewCalendarEventRepository(pool)
	donnaEventSvc := business.NewDonnaEventService(donnaEventRepo)
	donnaReminderSvc := business.NewDonnaReminderService(donnaReminderRepo)
	timelineSvc := business.NewTimelineService(business.TimelineServiceDeps{
		Providers: []business.TimelineProvider{
			business.NewGoogleTimelineProvider(calendarEventsRepo),
			business.NewMicrosoftICSTimelineProvider(calendarEventsRepo),
			business.NewDonnaEventTimelineProvider(donnaEventRepo),
			business.NewDonnaReminderTimelineProvider(donnaReminderRepo),
		},
	})
	notificationRepo := repository.NewNotificationRepository(pool)
	notificationSvc := business.NewNotificationService(
		notificationRepo,
		timelineSvc,
		business.NewNotificationPolicyResolver(),
	)

	actionRegistry := actions.NewRegistry(actions.Deps{
		Events:        donnaEventSvc,
		Reminders:     donnaReminderSvc,
		Tasks:         taskSvc,
		Timeline:      timelineSvc,
		Notifications: notificationSvc,
		Publisher:     actions.NoopPublisher{},
	})

	taskHandler := handler.NewTaskHandler(
		taskSvc,
		actionRegistry.CreateTask,
		actionRegistry.UpdateTask,
		actionRegistry.CompleteTask,
		actionRegistry.DeleteTask,
		taskLog,
	)

	noteLog := logFactory.Module(constant.ModuleNote)
	noteRepo := repository.NewNoteRepository(pool)
	noteSvc := business.NewNoteService(noteRepo)
	noteHandler := handler.NewNoteHandler(noteSvc, noteLog)

	timelineLog := logFactory.Module(constant.ModuleTimeline)
	timelineHandler := handler.NewTimelineHandler(actionRegistry.QueryTimeline, timelineLog)
	donnaEventHandler := handler.NewDonnaEventHandler(
		donnaEventSvc,
		actionRegistry.CreateEvent,
		actionRegistry.UpdateEvent,
		actionRegistry.DeleteEvent,
		timelineLog,
	)
	donnaReminderHandler := handler.NewDonnaReminderHandler(
		donnaReminderSvc,
		actionRegistry.CreateReminder,
		actionRegistry.UpdateReminder,
		actionRegistry.DeleteReminder,
		timelineLog,
	)

	notificationLog := logFactory.Module(constant.ModuleNotification)
	notificationHandler := handler.NewNotificationHandler(
		actionRegistry.GetNotifications,
		actionRegistry.MarkNotificationRead,
		actionRegistry.DismissNotification,
		notificationLog,
	)
	notificationScheduler := business.NewNotificationScheduler(notificationSvc, userRepo, notificationLog)

	chatLog := logFactory.Module(constant.ModuleChat)
	chatExecutor := chat.NewExecutor(chat.NewRuleBasedParser(), actionRegistry)
	tx := repository.NewTxManager(pool)
	conversationRepo := repository.NewConversationRepository(pool)
	messageRepo := repository.NewMessageRepository(pool)
	conversationSvc := business.NewConversationService(conversationRepo, messageRepo, tx)
	chatHandler := handler.NewChatHandler(chatExecutor, conversationSvc, userSvc, chatLog)

	notificationDispatcher := business.NewNotificationDispatcher(
		notificationRepo,
		notificationLog,
	)

	engine := router.New(router.Options{
		Environment:           cfg.App.Environment,
		CORSOrigins:           cfg.App.CORSOrigins,
		HTTPLogger:            httpLog,
		HealthHandler:         healthHandler,
		UserHandler:           userHandler,
		AuthHandler:           authParts.handler,
		MeHandler:             handler.NewMeHandler(userSvc, cfg.App.CookieSecure),
		CalendarHandler:       calendarParts.handler,
		IntegrationHandler:    calendarParts.integrationHandler,
		TaskHandler:           taskHandler,
		NoteHandler:           noteHandler,
		TimelineHandler:       timelineHandler,
		DonnaEventHandler:     donnaEventHandler,
		DonnaReminderHandler:  donnaReminderHandler,
		NotificationHandler:   notificationHandler,
		PushHandler:           nil, // Web Push disabled — Notification Center + Chat only
		ChatHandler:           chatHandler,
		TokenIssuer:           authParts.issuer,
	})

	srv := server.New(cfg.App.Addr, engine, appLog)

	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()
	go notificationScheduler.Run(runCtx)
	appLog.Info(ctx, "notification scheduler started")
	go notificationDispatcher.Run(runCtx)
	appLog.Info(ctx, "notification dispatcher started")
	if calendarParts.service != nil {
		jobsRepo := repository.NewSchedulerJobRepository(pool)
		platformJobs := []scheduler.Job{
			calendarsync.NewJob(calendarParts.service),
		}
		if err := scheduler.ValidateJobs(platformJobs); err != nil {
			return fmt.Errorf("scheduler: %w", err)
		}
		runner := scheduler.NewRunner(
			jobsRepo,
			logFactory.Module(constant.ModuleScheduler),
			platformJobs,
			scheduler.Options{},
		)
		go runner.Run(runCtx)
		appLog.Info(ctx, "scheduler runner started", "job_types", runner.RegisteredTypes())
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		runCancel()
		return err
	case sig := <-sigCh:
		appLog.Info(ctx, "shutdown signal received", constant.LogAttrSignal, sig.String())
		runCancel()
	case <-ctx.Done():
		appLog.Info(ctx, "shutdown context canceled", constant.LogAttrError, ctx.Err())
		runCancel()
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
	handler  *handler.AuthHandler
	service  *business.AuthService
	issuer   *session.Issuer
	state    *oauthstate.Manager
	sealKey  []byte
	accounts repository.ConnectedAccountRepository
	secrets  repository.CredentialSecretRepository
	tx       *repository.TxManager
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

	accounts := repository.NewConnectedAccountRepository(pool)
	secrets := repository.NewCredentialSecretRepository(pool)
	tx := repository.NewTxManager(pool)
	state := oauthstate.NewManager(cfg.App.JWTSecret)

	var googleLogin business.GoogleOAuthClient
	if cfg.API.GoogleOAuth.ClientID != "" && cfg.API.GoogleOAuth.ClientSecret != "" {
		client, err := googleoauth.NewClient(googleoauth.Config{
			ClientID:      cfg.API.GoogleOAuth.ClientID,
			ClientSecret:  cfg.API.GoogleOAuth.ClientSecret,
			RedirectURL:   cfg.API.GoogleOAuth.RedirectURL,
			AuthURL:       cfg.API.GoogleOAuth.AuthURL,
			TokenURL:      cfg.API.GoogleOAuth.TokenURL,
			UserInfoURL:   cfg.API.GoogleOAuth.UserInfoURL,
			Scopes:        cfg.API.GoogleOAuth.Scopes,
			Timeout:       cfg.API.GoogleOAuth.Timeout,
			OfflineAccess: true,
		})
		if err != nil {
			return authWire{}, fmt.Errorf("google login oauth client: %w", err)
		}
		googleLogin = client
		authLog.Info(context.Background(), "google login oauth client configured")
	} else {
		authLog.Warn(context.Background(), "google oauth not configured; /auth/google will reject requests")
	}

	var microsoftLogin business.MicrosoftOAuthClient
	if cfg.API.MicrosoftOAuth.ClientID != "" && cfg.API.MicrosoftOAuth.ClientSecret != "" {
		client, err := microsoftoauth.NewClient(microsoftoauth.Config{
			ClientID:     cfg.API.MicrosoftOAuth.ClientID,
			ClientSecret: cfg.API.MicrosoftOAuth.ClientSecret,
			RedirectURL:  cfg.API.MicrosoftOAuth.RedirectURL,
			AuthURL:      cfg.API.MicrosoftOAuth.AuthURL,
			TokenURL:     cfg.API.MicrosoftOAuth.TokenURL,
			GraphMeURL:   cfg.API.MicrosoftOAuth.GraphMeURL,
			Scopes:       cfg.API.MicrosoftOAuth.Scopes,
			Timeout:      cfg.API.MicrosoftOAuth.Timeout,
		})
		if err != nil {
			return authWire{}, fmt.Errorf("microsoft login oauth client: %w", err)
		}
		microsoftLogin = client
		authLog.Info(context.Background(), "microsoft login oauth client configured")
	} else {
		authLog.Warn(context.Background(), "microsoft oauth not configured; /auth/microsoft will reject requests")
	}

	authSvc := business.NewAuthService(business.AuthServiceDeps{
		Users:      userSvc,
		UserRepo:   userRepo,
		Identities: repository.NewAuthIdentityRepository(pool),
		Tx:         tx,
		Google:     googleLogin,
		Microsoft:  microsoftLogin,
		State:      state,
		Tokens:     tokenIssuer,
		Log:        authLog,
	})

	return authWire{
		handler: handler.NewAuthHandler(authSvc, authLog, handler.AuthHandlerConfig{
			FrontendSuccessURL: cfg.App.FrontendSuccessURL,
			AllowedOrigins:     cfg.App.CORSOrigins,
			CookieSecure:       cfg.App.CookieSecure,
			CookieMaxAge:       cfg.App.JWTExpiry,
			Tokens:             tokenIssuer,
			Users:              userSvc,
		}),
		service:  authSvc,
		issuer:   tokenIssuer,
		state:    state,
		sealKey:  sealKey,
		accounts: accounts,
		secrets:  secrets,
		tx:       tx,
	}, nil
}

func wireCalendar(
	cfg *config.Config,
	pool *pgxpool.Pool,
	auth authWire,
	calendarLog *logger.Logger,
) (calendarWire, error) {
	providers := map[string]calendarprovider.Provider{}
	tokens := map[string]calendarprovider.TokenRefresher{}

	var googleIntegration business.GoogleOAuthClient
	if cfg.API.GoogleOAuth.ClientID != "" && cfg.API.GoogleOAuth.ClientSecret != "" {
		client, err := googleoauth.NewClient(googleoauth.Config{
			ClientID:      cfg.API.GoogleOAuth.ClientID,
			ClientSecret:  cfg.API.GoogleOAuth.ClientSecret,
			RedirectURL:   cfg.API.GoogleOAuth.IntegrationRedirectURL,
			AuthURL:       cfg.API.GoogleOAuth.AuthURL,
			TokenURL:      cfg.API.GoogleOAuth.TokenURL,
			UserInfoURL:   cfg.API.GoogleOAuth.UserInfoURL,
			Scopes:        cfg.API.GoogleOAuth.IntegrationScopes,
			Timeout:       cfg.API.GoogleOAuth.Timeout,
			OfflineAccess: true,
		})
		if err != nil {
			return calendarWire{}, fmt.Errorf("google integration oauth client: %w", err)
		}
		googleIntegration = client
		calClient := googlecalendar.NewClient(googlecalendar.Config{
			BaseURL: constant.GoogleCalendarAPIBaseURL,
			Timeout: cfg.API.GoogleOAuth.Timeout,
		})
		providers[constant.AuthProviderGoogle] = googlecalendar.NewProvider(calClient)
		tokens[constant.AuthProviderGoogle] = business.GoogleTokenRefresher(client)
		calendarLog.Info(context.Background(), "google calendar provider configured")
	}

	var microsoftIntegration business.MicrosoftOAuthClient
	if cfg.API.MicrosoftOAuth.ClientID != "" && cfg.API.MicrosoftOAuth.ClientSecret != "" {
		client, err := microsoftoauth.NewClient(microsoftoauth.Config{
			ClientID:     cfg.API.MicrosoftOAuth.ClientID,
			ClientSecret: cfg.API.MicrosoftOAuth.ClientSecret,
			RedirectURL:  cfg.API.MicrosoftOAuth.IntegrationRedirectURL,
			AuthURL:      cfg.API.MicrosoftOAuth.AuthURL,
			TokenURL:     cfg.API.MicrosoftOAuth.TokenURL,
			GraphMeURL:   cfg.API.MicrosoftOAuth.GraphMeURL,
			Scopes:       cfg.API.MicrosoftOAuth.IntegrationScopes,
			Timeout:      cfg.API.MicrosoftOAuth.Timeout,
		})
		if err != nil {
			return calendarWire{}, fmt.Errorf("microsoft integration oauth client: %w", err)
		}
		microsoftIntegration = client
		msCal := microsoftcalendar.NewClient(microsoftcalendar.Config{
			BaseURL: constant.MicrosoftGraphAPIBaseURL,
			Timeout: cfg.API.MicrosoftOAuth.Timeout,
		})
		providers[constant.AuthProviderMicrosoft] = microsoftcalendar.NewProvider(msCal)
		tokens[constant.AuthProviderMicrosoft] = business.MicrosoftTokenRefresher(client.RefreshAsProvider)
		calendarLog.Info(context.Background(), "microsoft calendar provider configured")
	} else {
		calendarLog.Warn(context.Background(), "microsoft oauth not configured; outlook connect disabled")
	}

	// ICS is always available — provider-independent calendar feeds need no OAuth app.
	icsClient := icscalendar.NewClient(icscalendar.Config{
		UA: constant.ICSHTTPUserAgent,
	})
	providers[constant.AuthProviderICS] = icscalendar.NewProvider(icsClient)
	calendarLog.Info(context.Background(), "ics calendar provider configured")

	if len(providers) == 0 {
		calendarLog.Warn(context.Background(), "calendar module disabled; no calendar providers configured")
		return calendarWire{}, nil
	}

	sourcesRepo := repository.NewCalendarSourceRepository(pool)
	eventsRepo := repository.NewCalendarEventRepository(pool)

	svc := business.NewCalendarService(business.CalendarServiceDeps{
		Accounts:    auth.accounts,
		Secrets:     auth.secrets,
		Sources:     sourcesRepo,
		Events:      eventsRepo,
		DonnaEvents: repository.NewDonnaEventRepository(pool),
		SyncRuns:    repository.NewCalendarSyncRunRepository(pool),
		Jobs:        repository.NewSchedulerJobRepository(pool),
		Tx:          auth.tx,
		Providers:   providers,
		Tokens:      tokens,
		SealKey:     auth.sealKey,
		Log:         calendarLog,
	})

	out := calendarWire{
		handler: handler.NewCalendarHandler(svc, calendarLog),
		service: svc,
	}

	integrations := business.NewIntegrationService(business.IntegrationServiceDeps{
		Accounts:  auth.accounts,
		Secrets:   auth.secrets,
		Sources:   sourcesRepo,
		Events:    eventsRepo,
		Tx:        auth.tx,
		Google:    googleIntegration,
		Microsoft: microsoftIntegration,
		State:     auth.state,
		SealKey:   auth.sealKey,
		Log:       calendarLog,
	})
	out.integrations = integrations
	out.integrationHandler = handler.NewIntegrationHandler(
		integrations,
		calendarLog,
		cfg.App.IntegrationFrontendSuccessURL,
	)

	return out, nil
}

type calendarWire struct {
	handler            *handler.CalendarHandler
	service            *business.CalendarService
	integrations       *business.IntegrationService
	integrationHandler *handler.IntegrationHandler
}
