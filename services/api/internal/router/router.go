package router

import (
	"log/slog"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/handler"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Options configures the HTTP router.
type Options struct {
	Environment   string
	CORSOrigins   []string
	Logger        *slog.Logger
	HealthHandler *handler.HealthHandler
}

// New builds a Gin engine with middleware and /api/v1 routes.
// Auth and rate-limit middleware are intentionally omitted in M1;
// insert them here when M2/M10 land.
func New(opts Options) *gin.Engine {
	if opts.Environment == constant.EnvProduction || opts.Environment == constant.EnvStaging {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(
		middleware.Recovery(opts.Logger),
		middleware.RequestID(),
		middleware.RequestLogging(opts.Logger),
		middleware.CORS(opts.CORSOrigins),
	)

	v1 := r.Group(constant.APIPrefixV1)
	{
		v1.GET(constant.PathHealth, opts.HealthHandler.Health)
		v1.GET(constant.PathReady, opts.HealthHandler.Ready)
		v1.GET(constant.PathVersion, opts.HealthHandler.Version)
	}

	return r
}
