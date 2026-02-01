![gin-golang-boilerplate](https://socialify.git.ci/chuanghiduoc/gin-golang-boilerplate/image?font=Inter&forks=1&issues=1&language=1&name=1&owner=1&pattern=Plus&pulls=1&stargazers=1&theme=Auto)

# Backend Gin

Production-ready Gin boilerplate with Clean Architecture, PostgreSQL, Redis, JWT Auth, and Docker.

## Features

- **Clean Architecture**: 4-layer architecture (Domain, UseCase, Adapter, Infrastructure)
- **PostgreSQL + pgx/v5**: High-performance PostgreSQL driver
- **Redis Cache**: Caching and rate limiting support
- **SQLC**: Type-safe SQL code generation
- **JWT Authentication**: Access token with secure authentication
- **Role-Based Access Control**: Admin and User roles with middleware
- **Enterprise Response**: Standardized API responses with request_id, timestamp, i18n
- **File Upload**: Multi-driver storage (Local, AWS S3, Cloudflare R2)
- **i18n Support**: Internationalization (Vietnamese, English)
- **Swagger Documentation**: Auto-generated API docs
- **Docker Optimized**: Multi-stage build with UPX compression (~25MB image)
- **Hot Reload**: Development with Air (like nodemon)
- **Graceful Shutdown**: Production-ready server management
- **Structured Logging**: JSON logging with slog

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 16+
- Redis 7+
- Docker & Docker Compose

### Development Setup

```bash
# 1. Install development tools
make dev-setup

# 2. Start databases (PostgreSQL + Redis)
make dev-docker

# 3. Run migrations (new terminal)
make migrate-up

# 4. Start server with hot reload
make dev
```

### Production (Docker)

```bash
# Build optimized image (~25MB)
make docker-build

# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f api
```

### Access

- API: http://localhost:8080
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health

## API Response Format

All responses follow enterprise standards:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

Error response:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": [
      {
        "field": "email",
        "code": "REQUIRED",
        "message": "email is required"
      }
    ]
  },
  "meta": {
    "request_id": "req_abc123",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login and get JWT token |

### Users (Protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/me` | Get current user |
| GET | `/api/v1/users` | List all users |
| GET | `/api/v1/users/:id` | Get user by ID |
| POST | `/api/v1/users` | Create new user |
| PUT | `/api/v1/users/:id` | Update user |
| DELETE | `/api/v1/users/:id` | Delete user |

### Files (Protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/files` | Upload a file |
| GET | `/api/v1/files/me` | List my files |
| GET | `/api/v1/files/:id` | Get file by ID |
| DELETE | `/api/v1/files/:id` | Delete file |

### Admin (Admin Only)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/files` | List all files |
| GET | `/api/v1/admin/users` | List all users |

### Health
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

## Usage Examples

### Register
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Get current user (with i18n)
```bash
# English (default)
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <token>"

# Vietnamese
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <token>" \
  -H "Accept-Language: vi"
```

### Upload file
```bash
curl -X POST http://localhost:8080/api/v1/files \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/file.jpg" \
  -F "path=images"
```

## Project Structure

```
backend-gin/
├── cmd/
│   ├── api/main.go              # HTTP server entrypoint
│   └── migrate/main.go          # Migration CLI
├── internal/
│   ├── domain/                  # Layer 1: Domain (innermost)
│   │   ├── entity/              # Business entities
│   │   └── repository/          # Repository interfaces
│   ├── usecase/                 # Layer 2: Use Cases
│   │   ├── user/                # User service + DTOs
│   │   ├── auth/                # Auth service (JWT)
│   │   └── file/                # File service
│   ├── adapter/                 # Layer 3: Adapters
│   │   ├── handler/http/        # Gin handlers & middleware
│   │   └── repository/postgres/ # SQLC implementations
│   └── infrastructure/          # Layer 4: Infrastructure
│       ├── config/              # Environment config
│       ├── database/            # PostgreSQL connection
│       ├── cache/               # Redis cache
│       ├── logger/              # Structured logging
│       ├── storage/             # File storage (Local/S3/R2)
│       ├── i18n/                # Internationalization
│       └── server/              # Graceful shutdown
├── pkg/apperror/                # Application errors
├── db/
│   ├── migrations/              # SQL migration files
│   ├── queries/                 # SQLC query files
│   └── sqlc/                    # Generated code
├── docs/                        # Swagger documentation
├── scripts/                     # Utility scripts
│   └── bump-version.sh          # Version bump script
├── .air.toml                    # Hot reload config
├── Dockerfile                   # Optimized multi-stage build
├── docker-compose.yml           # Production compose
├── docker-compose.dev.yml       # Development compose
└── Makefile
```

## Available Commands

```bash
# Development
make dev-setup          # Install dev tools (air, sqlc, swag, migrate, lint)
make dev-docker         # Start PostgreSQL + Redis
make dev                # Run with hot reload

# Build
make build              # Build binaries
make docker-build       # Build Docker image (~25MB)

# Run
make run                # Run the API server

# Database
make migrate-up         # Run all up migrations
make migrate-down       # Run all down migrations
make migrate-down-one   # Rollback one migration
make sqlc               # Generate SQLC code

# Docker
make docker-up          # Start all containers
make docker-down        # Stop all containers
make docker-logs        # View logs

# Testing
make test               # Run all tests
make test-coverage      # Run tests with coverage

# Code Quality
make lint               # Run linter
make fmt                # Format code
make swagger            # Generate Swagger docs
make tidy               # Tidy dependencies

# Version Management
make version            # Show current version
make version-patch      # Bump patch (1.0.0 -> 1.0.1)
make version-minor      # Bump minor (1.0.0 -> 1.1.0)
make version-major      # Bump major (1.0.0 -> 2.0.0)
make release            # Release patch (bump + commit + tag + push)
make release-minor      # Release minor
make release-major      # Release major
```

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_HOST` | Server bind address | 0.0.0.0 |
| `SERVER_PORT` | Server port | 8080 |
| `GIN_MODE` | Gin mode (debug/release) | debug |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_USER` | Database user | postgres |
| `DB_PASSWORD` | Database password | postgres |
| `DB_NAME` | Database name | backend_gin |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `JWT_SECRET` | JWT signing secret | (required) |
| `JWT_ACCESS_EXPIRATION_MINUTES` | Access token expiration | 15 |
| `STORAGE_DRIVER` | Storage driver (local/s3/r2) | local |

## i18n (Internationalization)

Supported languages:
- English (en) - default
- Vietnamese (vi)

Set language via:
- Header: `Accept-Language: vi`
- Query: `?lang=vi`

## License

MIT
