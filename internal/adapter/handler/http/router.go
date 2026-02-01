package http

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"backend-gin/internal/adapter/handler/http/middleware"
	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
	"backend-gin/internal/infrastructure/config"
	"backend-gin/internal/infrastructure/logger"
	"backend-gin/internal/usecase/auth"
)

type Router struct {
	engine          *gin.Engine
	authHandler     *AuthHandler
	userHandler     *UserHandler
	fileHandler     *FileHandler
	healthHandler   *HealthHandler
	authService     auth.Service
	userRepo        repository.UserRepository
	redisClient     *redis.Client
	rateLimitConfig config.RateLimitConfig
	log             *logger.Logger
	mode            string
	uploadPath      string
	version         string
}

func NewRouter(
	authHandler *AuthHandler,
	userHandler *UserHandler,
	fileHandler *FileHandler,
	healthHandler *HealthHandler,
	authService auth.Service,
	userRepo repository.UserRepository,
	redisClient *redis.Client,
	rateLimitConfig config.RateLimitConfig,
	log *logger.Logger,
	mode string,
	uploadPath string,
	version string,
) *Router {
	gin.SetMode(mode)
	engine := gin.New()

	return &Router{
		engine:          engine,
		authHandler:     authHandler,
		userHandler:     userHandler,
		fileHandler:     fileHandler,
		healthHandler:   healthHandler,
		authService:     authService,
		userRepo:        userRepo,
		redisClient:     redisClient,
		rateLimitConfig: rateLimitConfig,
		log:             log,
		mode:            mode,
		uploadPath:      uploadPath,
		version:         version,
	}
}

func (r *Router) Setup() *gin.Engine {
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.Recovery(r.log))
	r.engine.Use(middleware.Logger(r.log))
	r.engine.Use(middleware.CORS())
	r.engine.Use(middleware.Security())
	r.engine.Use(middleware.I18n())

	if r.rateLimitConfig.Enabled && r.redisClient != nil {
		r.engine.Use(middleware.RateLimitWithConfig(r.redisClient, middleware.RateLimitConfig{
			Max:       r.rateLimitConfig.MaxRequests,
			WindowSec: r.rateLimitConfig.WindowSec,
			KeyPrefix: "ratelimit:global:",
			Message:   "too many requests, please try again later",
		}))
		r.log.Info("rate limiting enabled", "max", r.rateLimitConfig.MaxRequests, "window", r.rateLimitConfig.WindowSec)
	}

	r.engine.GET("/health", r.healthHandler.Health)

	// Root endpoint - show API info
	r.engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "Backend Gin API",
			"version": r.version,
			"health":  "/health",
			"docs":    "/swagger/index.html",
		})
	})

	// Swagger - disabled in production (release mode)
	if r.mode != "release" {
		r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		r.log.Info("swagger enabled", "url", "/swagger/index.html")
	} else {
		r.log.Info("swagger disabled in production mode")
	}

	r.engine.Static("/uploads", r.uploadPath)

	api := r.engine.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", r.authHandler.Register)
			authGroup.POST("/login", r.authHandler.Login)
			authGroup.POST("/refresh", r.authHandler.RefreshToken)
			authGroup.POST("/logout", middleware.Auth(r.authService), r.authHandler.Logout)
			authGroup.POST("/logout-all", middleware.Auth(r.authService), r.authHandler.LogoutAll)
		}

		usersGroup := api.Group("/users")
		usersGroup.Use(middleware.Auth(r.authService))
		{
			usersGroup.GET("/me", r.userHandler.GetMe)
			usersGroup.GET("", r.userHandler.ListUsers)
			usersGroup.GET("/:id", r.userHandler.GetUser)
			usersGroup.POST("", r.userHandler.CreateUser)
			usersGroup.PUT("/:id", r.userHandler.UpdateUser)
			usersGroup.DELETE("/:id", r.userHandler.DeleteUser)
		}

		filesGroup := api.Group("/files")
		filesGroup.Use(middleware.Auth(r.authService))
		{
			filesGroup.POST("", r.fileHandler.Upload)
			filesGroup.GET("/me", r.fileHandler.ListMyFiles)
			filesGroup.GET("/:id", r.fileHandler.GetFile)
			filesGroup.DELETE("/:id", r.fileHandler.DeleteFile)
		}

		adminFilesGroup := api.Group("/admin/files")
		adminFilesGroup.Use(middleware.Auth(r.authService))
		adminFilesGroup.Use(middleware.RequireAdmin(r.userRepo))
		{
			adminFilesGroup.GET("", r.fileHandler.ListAllFiles)
		}

		adminUsersGroup := api.Group("/admin/users")
		adminUsersGroup.Use(middleware.Auth(r.authService))
		adminUsersGroup.Use(middleware.RequireRole(r.userRepo, entity.RoleAdmin))
		{
			adminUsersGroup.GET("", r.userHandler.ListUsers)
		}
	}

	return r.engine
}
