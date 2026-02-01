package file

import (
	"context"
	"errors"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"backend-gin/internal/domain/entity"
	"backend-gin/internal/domain/repository"
	"backend-gin/internal/infrastructure/storage"
	"backend-gin/pkg/apperror"
)

var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"application/pdf": true,
	"text/plain":      true,
	"text/csv":        true,
	"application/json": true,
	"application/xml":  true,
	"application/zip":  true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
}

const maxFileSize = 10 * 1024 * 1024 // 10MB

type Service interface {
	Upload(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, path string) (*FileResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*FileResponse, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, req *ListFilesRequest) (*ListFilesResponse, error)
	List(ctx context.Context, req *ListFilesRequest) (*ListFilesResponse, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error
}

type service struct {
	fileRepo       repository.FileRepository
	storageManager *storage.Manager
}

func NewService(fileRepo repository.FileRepository, storageManager *storage.Manager) Service {
	return &service{
		fileRepo:       fileRepo,
		storageManager: storageManager,
	}
}

func (s *service) Upload(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, path string) (*FileResponse, error) {
	if file.Size == 0 {
		return nil, apperror.BadRequest("file is empty")
	}

	if file.Size > maxFileSize {
		return nil, apperror.BadRequest("file size exceeds the limit")
	}

	mimeType := file.Header.Get("Content-Type")
	if !allowedMimeTypes[mimeType] {
		return nil, apperror.BadRequest("invalid file type")
	}

	if path == "" {
		path = "uploads"
	}
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	storageDriver := s.storageManager.Default()
	result, err := storageDriver.Upload(ctx, file, path)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to upload file")
	}

	fileEntity := &entity.File{
		UserID:        userID,
		Filename:      result.Filename,
		OriginalName:  file.Filename,
		MimeType:      result.MimeType,
		Size:          result.Size,
		StorageDriver: storageDriver.Driver(),
		StoragePath:   result.StoragePath,
		URL:           result.URL,
	}

	createdFile, err := s.fileRepo.Create(ctx, fileEntity)
	if err != nil {
		_ = storageDriver.Delete(ctx, result.StoragePath)
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to save file record")
	}

	return ToFileResponse(createdFile), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*FileResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("file not found")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get file")
	}

	return ToFileResponse(file), nil
}

func (s *service) ListByUserID(ctx context.Context, userID uuid.UUID, req *ListFilesRequest) (*ListFilesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	offset := int32((req.Page - 1) * req.PageSize)
	limit := int32(req.PageSize)

	files, err := s.fileRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to list files")
	}

	total, err := s.fileRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to count files")
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListFilesResponse{
		Files:      ToFileResponses(files),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) List(ctx context.Context, req *ListFilesRequest) (*ListFilesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	offset := int32((req.Page - 1) * req.PageSize)
	limit := int32(req.PageSize)

	files, err := s.fileRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to list files")
	}

	total, err := s.fileRepo.Count(ctx)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to count files")
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &ListFilesResponse{
		Files:      ToFileResponses(files),
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("file not found")
		}
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get file")
	}

	if file.UserID != userID && !isAdmin {
		return apperror.Forbidden("you can only delete your own files")
	}

	storageDriver := s.storageManager.Get(file.StorageDriver)
	if err := storageDriver.Delete(ctx, file.StoragePath); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to delete file from storage")
	}

	if err := s.fileRepo.Delete(ctx, id); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to delete file record")
	}

	return nil
}
