package router

import (
	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/handler"
	"github.com/RandomThacker/donna/services/api/internal/logger"
	"github.com/RandomThacker/donna/services/api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Options configures the HTTP router.
type Options struct {
	Environment   string
	CORSOrigins   []string
	HTTPLogger    *logger.Logger
	HealthHandler *handler.HealthHandler
	UserHandler   *handler.UserHandler
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
	}

	return r
}
