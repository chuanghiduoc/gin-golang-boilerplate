package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend-gin/db/sqlc"
	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
)

type fileRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewFileRepository(pool *pgxpool.Pool) repository.FileRepository {
	return &fileRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *fileRepository) Create(ctx context.Context, file *entity.File) (*entity.File, error) {
	var url *string
	if file.URL != "" {
		url = &file.URL
	}

	result, err := r.queries.CreateFile(ctx, sqlc.CreateFileParams{
		UserID:        file.UserID,
		Filename:      file.Filename,
		OriginalName:  file.OriginalName,
		MimeType:      file.MimeType,
		Size:          file.Size,
		StorageDriver: string(file.StorageDriver),
		StoragePath:   file.StoragePath,
		URL:           url,
	})
	if err != nil {
		return nil, err
	}
	return toFileEntity(&result), nil
}

func (r *fileRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.File, error) {
	result, err := r.queries.GetFileByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toFileEntity(&result), nil
}

func (r *fileRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*entity.File, error) {
	results, err := r.queries.ListFilesByUserID(ctx, sqlc.ListFilesByUserIDParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return toFileEntities(results), nil
}

func (r *fileRepository) List(ctx context.Context, limit, offset int32) ([]*entity.File, error) {
	results, err := r.queries.ListFiles(ctx, sqlc.ListFilesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	return toFileEntities(results), nil
}

func (r *fileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteFile(ctx, id)
}

func (r *fileRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteFilesByUserID(ctx, userID)
}

func (r *fileRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.queries.CountFilesByUserID(ctx, userID)
}

func (r *fileRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountFiles(ctx)
}

func toFileEntity(f *sqlc.File) *entity.File {
	url := ""
	if f.URL != nil {
		url = *f.URL
	}

	return &entity.File{
		ID:            f.ID,
		UserID:        f.UserID,
		Filename:      f.Filename,
		OriginalName:  f.OriginalName,
		MimeType:      f.MimeType,
		Size:          f.Size,
		StorageDriver: entity.ParseStorageDriver(f.StorageDriver),
		StoragePath:   f.StoragePath,
		URL:           url,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
	}
}

func toFileEntities(files []sqlc.File) []*entity.File {
	entities := make([]*entity.File, len(files))
	for i, f := range files {
		entities[i] = toFileEntity(&f)
	}
	return entities
}
