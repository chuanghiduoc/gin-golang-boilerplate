package user

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"backend-gin/internal/domain/entity"
	"backend-gin/pkg/pagination"
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

func TestService_GetByID_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	userID := uuid.New()
	expectedUser := &entity.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  entity.RoleUser,
	}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	result, err := svc.GetByID(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedUser.Email, result.Email)
	assert.Equal(t, expectedUser.Name, result.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByID", ctx, userID).Return(nil, pgx.ErrNoRows)

	result, err := svc.GetByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestService_Create_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	req := &CreateUserRequest{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, pgx.ErrNoRows)

	createdUser := &entity.User{
		ID:    uuid.New(),
		Email: req.Email,
		Name:  req.Name,
		Role:  entity.RoleUser,
	}
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.User")).Return(createdUser, nil)

	result, err := svc.Create(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Email, result.Email)
	assert.Equal(t, req.Name, result.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_Create_EmailExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	req := &CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	existingUser := &entity.User{
		ID:    uuid.New(),
		Email: req.Email,
	}
	mockRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

	result, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user.email_exists")

	mockRepo.AssertExpectations(t)
}

func TestService_Update_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	userID := uuid.New()
	req := &UpdateUserRequest{
		Name: "Updated Name",
	}

	existingUser := &entity.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Old Name",
		Role:  entity.RoleUser,
	}
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)

	updatedUser := &entity.User{
		ID:    userID,
		Email: existingUser.Email,
		Name:  req.Name,
		Role:  existingUser.Role,
	}
	mockRepo.On("Update", ctx, mock.AnythingOfType("*entity.User")).Return(updatedUser, nil)

	result, err := svc.Update(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Name, result.Name)

	mockRepo.AssertExpectations(t)
}

func TestService_Update_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	userID := uuid.New()
	req := &UpdateUserRequest{
		Name: "Updated Name",
	}

	mockRepo.On("GetByID", ctx, userID).Return(nil, pgx.ErrNoRows)

	result, err := svc.Update(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestService_Delete_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	userID := uuid.New()

	existingUser := &entity.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
	}
	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Delete", ctx, userID).Return(nil)

	err := svc.Delete(ctx, userID)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_Delete_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	userID := uuid.New()

	mockRepo.On("GetByID", ctx, userID).Return(nil, pgx.ErrNoRows)

	err := svc.Delete(ctx, userID)

	assert.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestService_List_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	req := &ListUsersRequest{
		Request: pagination.Request{Page: 1, PageSize: 10},
	}

	users := []*entity.User{
		{ID: uuid.New(), Email: "user1@example.com", Name: "User 1", Role: entity.RoleUser},
		{ID: uuid.New(), Email: "user2@example.com", Name: "User 2", Role: entity.RoleUser},
	}
	var total int64 = 2

	mockRepo.On("List", ctx, int32(10), int32(0)).Return(users, nil)
	mockRepo.On("Count", ctx).Return(total, nil)

	result, err := svc.List(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, total, result.Meta.Total)

	mockRepo.AssertExpectations(t)
}

func TestService_List_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	req := &ListUsersRequest{
		Request: pagination.Request{Page: 1, PageSize: 10},
	}

	mockRepo.On("List", ctx, int32(10), int32(0)).Return(nil, errors.New("database error"))

	result, err := svc.List(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestService_List_DefaultPagination(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewService(mockRepo, 10)

	ctx := context.Background()
	req := &ListUsersRequest{}

	users := []*entity.User{}
	var total int64 = 0

	mockRepo.On("List", ctx, int32(10), int32(0)).Return(users, nil)
	mockRepo.On("Count", ctx).Return(total, nil)

	result, err := svc.List(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Meta.Page)
	assert.Equal(t, 10, result.Meta.PageSize)

	mockRepo.AssertExpectations(t)
}
