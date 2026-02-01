package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	_ "backend-gin/docs"
	httphandler "backend-gin/internal/adapter/handler/http"
	"backend-gin/internal/adapter/repository/postgres"
	"backend-gin/internal/domain/entity"
	"backend-gin/internal/infrastructure/cache"
	"backend-gin/internal/infrastructure/config"
	"backend-gin/internal/infrastructure/database"
	"backend-gin/internal/infrastructure/i18n"
	"backend-gin/internal/infrastructure/logger"
	"backend-gin/internal/infrastructure/server"
	"backend-gin/internal/infrastructure/storage"
	"backend-gin/internal/usecase/auth"
	"backend-gin/internal/usecase/file"
	"backend-gin/internal/usecase/user"
)

// @title Backend Gin API
// @version 1.0
// @description Production-ready Gin boilerplate with Clean Architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.Load()

	appLogger := logger.New(cfg.Log.Level)

	if err := i18n.Init(""); err != nil {
		appLogger.Warn("failed to load i18n translations", "error", err)
	}

	pool, err := database.NewPostgresPool(&cfg.Database, appLogger)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	var redisClient *cache.RedisClient
	if cfg.Redis.Enabled {
		redisClient, err = cache.NewRedisClient(&cfg.Redis, appLogger)
		if err != nil {
			appLogger.Warn("failed to connect to Redis, continuing without cache", "error", err)
		} else {
			defer redisClient.Close()
		}
	}

	storageManager, err := initStorage(cfg, appLogger)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	app := initializeApp(pool, cfg, appLogger, storageManager, redisClient)

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := server.New(addr, app.Setup(), appLogger)

	if err := srv.Run(); err != nil {
		appLogger.Error("server error", "error", err)
	}
}

func initStorage(cfg *config.Config, log *logger.Logger) (*storage.Manager, error) {
	defaultDriver := entity.ParseStorageDriver(cfg.Storage.Driver)
	manager := storage.NewManager(defaultDriver)

	localStorage, err := storage.NewLocalStorage(storage.LocalConfig{
		BasePath: cfg.Storage.Local.BasePath,
		BaseURL:  cfg.Storage.Local.BaseURL,
	})
	if err != nil {
		return nil, err
	}
	manager.Register(localStorage)

	if cfg.Storage.S3.AccessKeyID != "" && cfg.Storage.S3.Bucket != "" {
		s3Storage, err := storage.NewS3Storage(context.Background(), storage.S3Config{
			Region:          cfg.Storage.S3.Region,
			AccessKeyID:     cfg.Storage.S3.AccessKeyID,
			SecretAccessKey: cfg.Storage.S3.SecretAccessKey,
			Bucket:          cfg.Storage.S3.Bucket,
			BaseURL:         cfg.Storage.S3.BaseURL,
			PathPrefix:      cfg.Storage.S3.PathPrefix,
			Endpoint:        cfg.Storage.S3.Endpoint,
		})
		if err != nil {
			log.Warn("failed to initialize S3 storage", "error", err)
		} else {
			manager.Register(s3Storage)
		}
	}

	if cfg.Storage.R2.AccountID != "" && cfg.Storage.R2.Bucket != "" {
		r2Storage, err := storage.NewR2Storage(context.Background(), storage.R2Config{
			AccountID:       cfg.Storage.R2.AccountID,
			AccessKeyID:     cfg.Storage.R2.AccessKeyID,
			SecretAccessKey: cfg.Storage.R2.SecretAccessKey,
			Bucket:          cfg.Storage.R2.Bucket,
			BaseURL:         cfg.Storage.R2.BaseURL,
			PathPrefix:      cfg.Storage.R2.PathPrefix,
		})
		if err != nil {
			log.Warn("failed to initialize R2 storage", "error", err)
		} else {
			manager.Register(r2Storage)
		}
	}

	log.Info("storage initialized", "driver", defaultDriver)
	return manager, nil
}

func initializeApp(pool *pgxpool.Pool, cfg *config.Config, log *logger.Logger, storageManager *storage.Manager, redisClient *cache.RedisClient) *httphandler.Router {
	userRepo := postgres.NewUserRepository(pool)
	fileRepo := postgres.NewFileRepository(pool)

	var rClient *redis.Client
	if redisClient != nil {
		rClient = redisClient.Client()
	}

	authService := auth.NewService(userRepo, rClient, cfg.JWT)
	userService := user.NewService(userRepo)
	fileService := file.NewService(fileRepo, storageManager)

	authHandler := httphandler.NewAuthHandler(authService)
	userHandler := httphandler.NewUserHandler(userService)
	fileHandler := httphandler.NewFileHandler(fileService)
	healthHandler := httphandler.NewHealthHandler(pool)

	router := httphandler.NewRouter(
		authHandler,
		userHandler,
		fileHandler,
		healthHandler,
		authService,
		userRepo,
		rClient,
		cfg.RateLimit,
		log,
		cfg.Server.Mode,
		cfg.Storage.Local.BasePath,
		Version,
	)

	return router
}
