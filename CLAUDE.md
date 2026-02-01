# CLAUDE.md - AI Coding Guidelines

## Project Overview

Production-ready Gin boilerplate following Clean Architecture principles.

## Tech Stack

- **Go**: 1.25.x
- **Framework**: Gin v1.11.0
- **Database**: PostgreSQL 16 + pgx/v5
- **SQL Generator**: SQLC
- **Migrations**: golang-migrate
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Docs**: Swagger (swaggo)
- **Logging**: slog (built-in)

## Project Structure

```
backend-gin/
├── cmd/                         # Application entrypoints
│   ├── api/
│   │   ├── main.go              # HTTP server
│   │   └── version.go           # Version info (injected at build)
│   └── migrate/main.go          # Migration CLI
├── scripts/                     # Utility scripts
│   └── bump-version.sh          # Semver bump script
├── internal/                    # Private application code
│   ├── domain/                  # Layer 1: Domain (innermost)
│   │   ├── entity/              # Business entities
│   │   └── repository/          # Repository interfaces
│   ├── usecase/                 # Layer 2: Use Cases
│   │   ├── user/                # User service + DTOs
│   │   └── auth/                # Auth service (JWT)
│   ├── adapter/                 # Layer 3: Adapters
│   │   ├── repository/postgres/ # SQLC implementations
│   │   └── handler/http/        # Gin handlers + middleware
│   └── infrastructure/          # Layer 4: Infrastructure
│       ├── config/              # Environment config
│       ├── database/            # PostgreSQL connection
│       ├── logger/              # Structured logging
│       └── server/              # Graceful shutdown
├── pkg/                         # Public packages
│   └── apperror/                # Application errors
├── db/                          # Database files
│   ├── migrations/              # SQL migrations
│   ├── queries/                 # SQLC queries
│   └── sqlc/                    # Generated code
└── docs/                        # Swagger documentation
```

## Clean Architecture Rules

### Dependency Rule
Dependencies point INWARD only:
- Domain → (no dependencies)
- Use Case → Domain
- Adapter → Use Case, Domain
- Infrastructure → Use Case, Domain, Adapter

### Layer Responsibilities

1. **Domain** (`internal/domain/`)
   - Contains business entities and repository interfaces
   - NO external dependencies
   - Pure Go structs and interfaces

2. **Use Case** (`internal/usecase/`)
   - Contains business logic
   - Depends only on Domain layer
   - DTOs for input/output

3. **Adapter** (`internal/adapter/`)
   - Implements repository interfaces
   - HTTP handlers
   - Converts between external and internal formats

4. **Infrastructure** (`internal/infrastructure/`)
   - Technical concerns (config, database, logging)
   - Framework-specific code

## Coding Conventions

### File Naming
- Use snake_case for file names: `user_repository.go`
- Suffix interfaces: `_repository.go`, `_service.go`
- Suffix implementations: `user_handler.go`

### Package Naming
- Use single lowercase word: `user`, `auth`, `postgres`
- Avoid stutter: `user.User` not `user.UserModel`

### Error Handling
```go
// Use apperror for application errors
return nil, apperror.NotFound("user not found")
return nil, apperror.Wrap(err, 500, "failed to create user")

// Check error type
if apperror.IsNotFound(err) {
    // handle not found
}
```

### DTOs
- Request DTOs: `CreateUserRequest`, `LoginRequest`
- Response DTOs: `UserResponse`, `AuthResponse`
- Use binding tags for validation

### Repository Pattern
```go
// Interface in domain layer
type UserRepository interface {
    Create(ctx context.Context, user *entity.User) (*entity.User, error)
    GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

// Implementation in adapter layer
type userRepository struct {
    pool    *pgxpool.Pool
    queries *sqlc.Queries
}
```

## SQLC Guidelines

### Query Naming
```sql
-- name: GetUserByID :one
-- name: ListUsers :many
-- name: CreateUser :one
-- name: UpdateUser :one
-- name: DeleteUser :exec
-- name: CountUsers :one
```

### Running SQLC
```bash
make sqlc
# or
sqlc generate
```

## API Patterns

### Response Format
```json
// Success
{
    "success": true,
    "data": { ... }
}

// Error
{
    "success": false,
    "error": {
        "code": 400,
        "message": "validation error"
    }
}

// Paginated
{
    "success": true,
    "data": [...],
    "meta": {
        "total": 100,
        "page": 1,
        "page_size": 10,
        "total_pages": 10
    }
}
```

### HTTP Status Codes
- 200: Success
- 201: Created
- 204: No Content
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 409: Conflict
- 422: Validation Error
- 500: Internal Server Error

## Testing Guidelines

### Unit Tests
- Test files: `*_test.go`
- Use table-driven tests
- Mock interfaces, not implementations

### Running Tests
```bash
make test              # Run all tests
make test-coverage     # With coverage report
```

## Git Commit Conventions

Format: `type(scope): description`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code refactoring
- `docs`: Documentation
- `test`: Tests
- `chore`: Maintenance

Examples:
```
feat(auth): add JWT refresh token
fix(user): handle duplicate email error
refactor(handler): extract response helpers
```

## Common Commands

```bash
# Development
make run              # Run locally
make dev              # Hot reload (requires air)

# Database
make migrate-up       # Run migrations
make migrate-down     # Rollback migrations
make sqlc             # Generate SQLC code

# Docker
make docker-up        # Start containers
make docker-down      # Stop containers
make docker-logs      # View logs

# Code Quality
make lint             # Run linter
make fmt              # Format code
make test             # Run tests

# Version Management
make version          # Show current version
make version-patch    # Bump patch (1.0.0 -> 1.0.1)
make version-minor    # Bump minor (1.0.0 -> 1.1.0)
make version-major    # Bump major (1.0.0 -> 2.0.0)
make release          # Release patch (bump + commit + tag + push)
make release-minor    # Release minor version
make release-major    # Release major version
```

## Environment Variables

Required:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- `JWT_SECRET`

Optional:
- `SERVER_HOST` (default: 0.0.0.0)
- `SERVER_PORT` (default: 8080)
- `GIN_MODE` (default: debug)
- `LOG_LEVEL` (default: debug)
