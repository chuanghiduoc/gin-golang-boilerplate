package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	Log       LogConfig
	Storage   StorageConfig
}

type ServerConfig struct {
	Host string
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxConnections     int
	MaxIdleConnections int
	MaxLifetime        time.Duration
}

type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	Enabled      bool
}

type JWTConfig struct {
	Secret               string
	AccessExpiration     time.Duration
	RefreshExpiration    time.Duration
	RefreshTokenLength   int
}

type RateLimitConfig struct {
	Enabled     bool
	MaxRequests int
	WindowSec   int
}

type LogConfig struct {
	Level string
}

type StorageConfig struct {
	Driver string
	Local  LocalStorageConfig
	S3     S3StorageConfig
	R2     R2StorageConfig
}

type LocalStorageConfig struct {
	BasePath string
	BaseURL  string
}

type S3StorageConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	PathPrefix      string
	Endpoint        string
}

type R2StorageConfig struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	PathPrefix      string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnv("DB_PORT", "5432"),
			User:               getEnv("DB_USER", "postgres"),
			Password:           getEnv("DB_PASSWORD", "postgres"),
			DBName:             getEnv("DB_NAME", "backend_gin"),
			SSLMode:            getEnv("DB_SSLMODE", "disable"),
			MaxConnections:     getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConnections: getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
			MaxLifetime:        time.Duration(getEnvAsInt("DB_MAX_LIFETIME_MINUTES", 5)) * time.Minute,
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			Enabled:      getEnvAsBool("REDIS_ENABLED", true),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			AccessExpiration:   time.Duration(getEnvAsInt("JWT_ACCESS_EXPIRATION_MINUTES", 15)) * time.Minute,
			RefreshExpiration:  time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,
			RefreshTokenLength: getEnvAsInt("JWT_REFRESH_TOKEN_LENGTH", 64),
		},
		RateLimit: RateLimitConfig{
			Enabled:     getEnvAsBool("RATE_LIMIT_ENABLED", true),
			MaxRequests: getEnvAsInt("RATE_LIMIT_MAX_REQUESTS", 100),
			WindowSec:   getEnvAsInt("RATE_LIMIT_WINDOW_SEC", 60),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
		},
		Storage: StorageConfig{
			Driver: getEnv("STORAGE_DRIVER", "local"),
			Local: LocalStorageConfig{
				BasePath: getEnv("STORAGE_LOCAL_PATH", "./uploads"),
				BaseURL:  getEnv("STORAGE_LOCAL_URL", "http://localhost:8080/uploads"),
			},
			S3: S3StorageConfig{
				Region:          getEnv("AWS_REGION", "us-east-1"),
				AccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
				SecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
				Bucket:          getEnv("AWS_S3_BUCKET", ""),
				BaseURL:         getEnv("AWS_S3_BASE_URL", ""),
				PathPrefix:      getEnv("AWS_S3_PATH_PREFIX", ""),
				Endpoint:        getEnv("AWS_S3_ENDPOINT", ""),
			},
			R2: R2StorageConfig{
				AccountID:       getEnv("R2_ACCOUNT_ID", ""),
				AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
				SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
				Bucket:          getEnv("R2_BUCKET", ""),
				BaseURL:         getEnv("R2_BASE_URL", ""),
				PathPrefix:      getEnv("R2_PATH_PREFIX", ""),
			},
		},
	}
}

func (c *DatabaseConfig) DSN() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.DBName + "?sslmode=" + c.SSLMode
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
