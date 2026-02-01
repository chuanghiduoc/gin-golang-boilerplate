package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"backend-gin/internal/domain/entity"
	"backend-gin/internal/infrastructure/config"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role) (*entity.User, error) {
	args := m.Called(ctx, id, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int32) ([]*entity.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestService_Register_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:             "test-secret",
		AccessExpiration:   15 * time.Minute,
		RefreshExpiration:  7 * 24 * time.Hour,
		RefreshTokenLength: 64,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	ctx := context.Background()
	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, pgx.ErrNoRows)

	createdUser := &entity.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: "hashed",
		Name:         req.Name,
		Role:         entity.RoleUser,
	}
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(createdUser, nil)

	result, err := svc.Register(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, req.Email, result.User.Email)
	assert.Equal(t, req.Name, result.User.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_Register_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: 15 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	ctx := context.Background()
	req := &RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	existingUser := &entity.User{
		ID:    uuid.New(),
		Email: req.Email,
	}
	mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	result, err := svc.Register(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already registered")

	mockRepo.AssertExpectations(t)
}

func TestService_Login_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:             "test-secret",
		AccessExpiration:   15 * time.Minute,
		RefreshExpiration:  7 * 24 * time.Hour,
		RefreshTokenLength: 64,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	ctx := context.Background()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Role:         entity.RoleUser,
	}
	mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	result, err := svc.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.Equal(t, "Bearer", result.TokenType)
	assert.Equal(t, req.Email, result.User.Email)

	mockRepo.AssertExpectations(t)
}

func TestService_Login_InvalidEmail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: 15 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	ctx := context.Background()
	req := &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, pgx.ErrNoRows)

	result, err := svc.Login(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestService_Login_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: 15 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	ctx := context.Background()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "wrong_password",
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         "Test User",
		Role:         entity.RoleUser,
	}
	mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	result, err := svc.Login(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestService_ValidateToken_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: 15 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig).(*service)

	user := &entity.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  entity.RoleUser,
	}

	token, err := svc.generateAccessToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestService_ValidateToken_Invalid(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: 15 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	claims, err := svc.ValidateToken("invalid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_ValidateToken_Expired(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: -1 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig).(*service)

	user := &entity.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
		Role:  entity.RoleUser,
	}

	token, err := svc.generateAccessToken(user)
	assert.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_Register_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	jwtConfig := config.JWTConfig{
		Secret:           "test-secret",
		AccessExpiration: 15 * time.Minute,
	}

	svc := NewService(mockRepo, nil, jwtConfig)

	ctx := context.Background()
	req := &RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("database error"))

	result, err := svc.Register(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}
