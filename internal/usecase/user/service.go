package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
	"backend-gin/pkg/apperror"
	"backend-gin/pkg/pagination"
)

const (
	pgUniqueViolation = "23505"
)

type Service interface {
	Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	List(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	userRepo   repository.UserRepository
	bcryptCost int
}

func NewService(userRepo repository.UserRepository, bcryptCost int) Service {
	return &service{
		userRepo:   userRepo,
		bcryptCost: bcryptCost,
	}
}

func (s *service) Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	// Early check for better UX (avoid bcrypt cost if email exists)
	// Note: This check alone is NOT sufficient due to race conditions
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to check existing user")
	}
	if existingUser != nil {
		return nil, apperror.ConflictI18n("user.email_exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
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
		// Handle unique constraint violation (race condition protection)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return nil, apperror.ConflictI18n("user.email_exists")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to create user")
	}

	return ToUserResponse(createdUser), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFoundI18n("user.not_found")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	return ToUserResponse(user), nil
}

func (s *service) List(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	users, err := s.userRepo.List(ctx, req.Limit(), req.Offset())
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to list users")
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to count users")
	}

	return &ListUsersResponse{
		Users: ToUserResponses(users),
		Meta:  pagination.NewResult(total, &req.Request),
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error) {
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFoundI18n("user.not_found")
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
			return apperror.NotFoundI18n("user.not_found")
		}
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get user")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to delete user")
	}

	return nil
}
