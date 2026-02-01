package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
	"backend-gin/pkg/apperror"
)

type Service interface {
	Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	List(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to check existing user")
	}
	if existingUser != nil {
		return nil, apperror.Conflict("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to hash password")
	}

	user := &entity.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to create user")
	}

	return ToUserResponse(createdUser), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	return ToUserResponse(user), nil
}

func (s *service) List(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	offset := int32((req.Page - 1) * req.PageSize)
	limit := int32(req.PageSize)

	users, err := s.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to list users")
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to count users")
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListUsersResponse{
		Users:      ToUserResponses(users),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error) {
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("user not found")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	existingUser.Name = req.Name

	updatedUser, err := s.userRepo.Update(ctx, existingUser)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to update user")
	}

	return ToUserResponse(updatedUser), nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("user not found")
		}
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to delete user")
	}

	return nil
}
