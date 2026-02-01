package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"backend-gin/internal/domain/entity"
)

type FileRepository interface {
	Create(ctx context.Context, file *entity.File) (*entity.File, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.File, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.File, error)
	List(ctx context.Context, limit, offset int32) ([]*entity.File, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
	Count(ctx context.Context) (int64, error)

	// Transaction methods
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CreateTx(ctx context.Context, tx pgx.Tx, file *entity.File) (*entity.File, error)
	DeleteTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}
