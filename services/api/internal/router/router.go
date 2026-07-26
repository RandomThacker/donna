package router

import (
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/handler"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/RandomThacker/donna/services/api/internal/session"
	"github.com/gin-gonic/gin"
)

// Options configures the HTTP router.
type Options struct {
	Environment        string
	CORSOrigins        []string
	HTTPLogger         *logger.Logger
	HealthHandler      *handler.HealthHandler
	UserHandler        *handler.UserHandler
	AuthHandler        *handler.AuthHandler
	MeHandler          *handler.MeHandler
	CalendarHandler    *handler.CalendarHandler
	IntegrationHandler *handler.IntegrationHandler
	TokenIssuer        *session.Issuer
}

// New builds a Gin engine with middleware and /api/v1 routes.
func New(opts Options) *gin.Engine {
	if opts.Environment == constant.EnvProduction || opts.Environment == constant.EnvStaging {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(
		middleware.Recovery(opts.HTTPLogger),
		middleware.RequestID(),
		middleware.RequestLogging(opts.HTTPLogger),
		middleware.CORS(opts.CORSOrigins),
	)

	v1 := r.Group(constant.APIPrefixV1)
	{
		v1.GET(constant.PathHealth, opts.HealthHandler.Health)
		v1.GET(constant.PathReady, opts.HealthHandler.Ready)
		v1.GET(constant.PathVersion, opts.HealthHandler.Version)

		if opts.UserHandler != nil {
			v1.POST(constant.PathUsers, opts.UserHandler.Create)
			v1.GET(constant.PathUsers, opts.UserHandler.GetByEmail)
			v1.GET(constant.PathUserByID, opts.UserHandler.GetByID)
			v1.PATCH(constant.PathUserByID, opts.UserHandler.Update)
			v1.DELETE(constant.PathUserByID, opts.UserHandler.SoftDelete)
		}

		if opts.AuthHandler != nil {
			v1.GET(constant.PathAuthGoogle, opts.AuthHandler.BeginGoogle)
			v1.GET(constant.PathAuthGoogleCallback, opts.AuthHandler.GoogleCallback)
			v1.GET(constant.PathAuthMicrosoft, opts.AuthHandler.BeginMicrosoft)
			v1.GET(constant.PathAuthMicrosoftCallback, opts.AuthHandler.MicrosoftCallback)
			v1.POST(constant.PathAuthLogout, opts.AuthHandler.Logout)
		}

		if opts.MeHandler != nil && opts.TokenIssuer != nil {
			v1.GET(constant.PathMe, middleware.RequireAuth(opts.TokenIssuer), opts.MeHandler.Me)
		}

		if opts.CalendarHandler != nil && opts.TokenIssuer != nil {
			auth := middleware.RequireAuth(opts.TokenIssuer)
			v1.POST(constant.PathCalendarSync, auth, opts.CalendarHandler.SyncSources)
			v1.POST(constant.PathCalendarSyncEnsure, auth, opts.CalendarHandler.EnsureFreshSources)
			v1.GET(constant.PathCalendarSources, auth, opts.CalendarHandler.ListSources)
			v1.POST(constant.PathCalendarEventsSync, auth, opts.CalendarHandler.SyncEvents)
			v1.GET(constant.PathCalendarEvents, auth, opts.CalendarHandler.ListEvents)
		}

		if opts.IntegrationHandler != nil && opts.TokenIssuer != nil {
			auth := middleware.RequireAuth(opts.TokenIssuer)
			v1.GET(constant.PathIntegrations, auth, opts.IntegrationHandler.ListConnectedAccounts)
			v1.GET(constant.PathIntegrationsGoogle, auth, opts.IntegrationHandler.BeginGoogleConnect)
			v1.POST(constant.PathIntegrationsGoogle, auth, opts.IntegrationHandler.BeginGoogleConnect)
			v1.GET(constant.PathIntegrationsGoogleCallback, opts.IntegrationHandler.GoogleCallback)
			v1.GET(constant.PathIntegrationsMicrosoft, auth, opts.IntegrationHandler.BeginMicrosoftConnect)
			v1.POST(constant.PathIntegrationsMicrosoft, auth, opts.IntegrationHandler.BeginMicrosoftConnect)
			// Callback verifies signed state (includes user id); no session cookie required.
			v1.GET(constant.PathIntegrationsMicrosoftCallback, opts.IntegrationHandler.MicrosoftCallback)
			v1.GET(constant.PathIntegrationsICS, auth, opts.IntegrationHandler.ListICS)
			v1.POST(constant.PathIntegrationsICS, auth, opts.IntegrationHandler.ConnectICS)
			v1.PATCH(constant.PathIntegrationsICSByID, auth, opts.IntegrationHandler.UpdateICS)
			v1.DELETE(constant.PathIntegrationsICSByID, auth, opts.IntegrationHandler.DeleteICS)
			v1.POST(constant.PathIntegrationsICSSync, auth, opts.IntegrationHandler.SyncICS)
			v1.DELETE(constant.PathIntegrationsByID, auth, opts.IntegrationHandler.Disconnect)
		}
	}

	return r
}
