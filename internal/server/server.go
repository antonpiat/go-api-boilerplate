package server

import (
	"net/http"

	"github.com/antonpiat/go-api-boilerplate/internal/auth"
	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/antonpiat/go-api-boilerplate/internal/httpx"
	"github.com/antonpiat/go-api-boilerplate/internal/middleware"
	"github.com/antonpiat/go-api-boilerplate/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/antonpiat/go-api-boilerplate/docs"
)

type Dependencies struct {
	Config      *config.Config
	Logger      *zap.Logger
	Ready       ReadyChecker
	AuthHandler *auth.Handler
	AuthService *auth.Service
	UserHandler *user.Handler
}

func New(deps Dependencies) *gin.Engine {
	cfg := deps.Config
	if gin.Mode() != gin.TestMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Recovery(deps.Logger),
		middleware.Logger(deps.Logger),
		middleware.CORS(cfg.CORS),
		middleware.MaxBodySize(cfg.Server.MaxBodyBytes),
	)
	if cfg.Metrics.Enabled {
		r.Use(middleware.Metrics())
	}

	health := NewHealthHandler(deps.Ready)
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})
	r.GET("/health/live", health.Live)
	r.GET("/health/ready", health.Ready)

	if cfg.Metrics.Enabled {
		r.GET(cfg.Metrics.Path, gin.WrapH(promhttp.Handler()))
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.NoRoute(func(c *gin.Context) {
		httpx.Fail(c, httpx.NotFound("route not found"))
	})

	v1 := r.Group("/api/v1")
	{
		publicAuth := v1.Group("/auth")
		deps.AuthHandler.RegisterPublic(publicAuth)

		protectedAuth := v1.Group("/auth")
		protectedAuth.Use(middleware.JWT(deps.AuthService))
		deps.AuthHandler.RegisterProtected(protectedAuth)

		users := v1.Group("/users")
		users.Use(middleware.JWT(deps.AuthService))
		deps.UserHandler.Register(users)

		adminUsers := users.Group("")
		adminUsers.Use(middleware.RequireRole(user.RoleAdmin))
		deps.UserHandler.RegisterAdmin(adminUsers)
	}

	return r
}

func NewHTTPServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeoutDuration(),
		WriteTimeout: cfg.Server.WriteTimeoutDuration(),
		IdleTimeout:  cfg.Server.IdleTimeoutDuration(),
	}
}
