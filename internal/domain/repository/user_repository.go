package repository

import (
	"context"

	"github.com/google/uuid"

	"backend-gin/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}
