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
	"backend-gin/internal/infrastructure/config"
	"backend-gin/internal/infrastructure/storage"
	"backend-gin/pkg/apperror"
	"backend-gin/pkg/filetype"
	"backend-gin/pkg/pagination"
)

type Service interface {
	Upload(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, path string) (*FileResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*FileResponse, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, req *ListFilesRequest) (*ListFilesResponse, error)
	List(ctx context.Context, req *ListFilesRequest) (*ListFilesResponse, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error
}

type service struct {
	fileRepo         repository.FileRepository
	storageManager   *storage.Manager
	maxFileSize      int64
	allowedMimeTypes map[string]bool
}

func NewService(fileRepo repository.FileRepository, storageManager *storage.Manager, cfg config.FileConfig) Service {
	// Build allowed mime types map
	allowedTypes := make(map[string]bool)
	for _, t := range cfg.AllowedMimeTypes {
		allowedTypes[t] = true
	}

	return &service{
		fileRepo:         fileRepo,
		storageManager:   storageManager,
		maxFileSize:      cfg.MaxSize,
		allowedMimeTypes: allowedTypes,
	}
}

func (s *service) Upload(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader, path string) (*FileResponse, error) {
	if file.Size == 0 {
		return nil, apperror.BadRequestI18n("file.empty_file")
	}

	if file.Size > s.maxFileSize {
		return nil, apperror.BadRequestI18n("file.size_exceeded")
	}

	// Get claimed MIME type from header
	claimedMimeType := file.Header.Get("Content-Type")
	if !s.allowedMimeTypes[claimedMimeType] {
		return nil, apperror.BadRequestI18n("file.invalid_type")
	}

	// Validate actual content matches claimed type (security: prevent MIME type spoofing)
	valid, detectedType, err := filetype.ValidateContent(file, claimedMimeType)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to validate file content")
	}
	if !valid {
		return nil, apperror.BadRequestI18nWithParams("file.content_mismatch", map[string]string{
			"claimed":  claimedMimeType,
			"detected": detectedType,
		})
	}

	if path == "" {
		path = "uploads"
	}
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	// Begin transaction
	tx, err := s.fileRepo.BeginTx(ctx)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Upload to storage
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

	// Create DB record within transaction
	createdFile, err := s.fileRepo.CreateTx(ctx, tx, fileEntity)
	if err != nil {
		// Rollback: delete uploaded file from storage
		_ = storageDriver.Delete(ctx, result.StoragePath)
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to save file record")
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		// Rollback: delete uploaded file from storage
		_ = storageDriver.Delete(ctx, result.StoragePath)
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to commit transaction")
	}

	return ToFileResponse(createdFile), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*FileResponse, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFoundI18n("file.not_found")
		}
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get file")
	}

	return ToFileResponse(file), nil
}

func (s *service) ListByUserID(ctx context.Context, userID uuid.UUID, req *ListFilesRequest) (*ListFilesResponse, error) {
	files, err := s.fileRepo.ListByUserID(ctx, userID, req.Limit(), req.Offset())
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to list files")
	}

	total, err := s.fileRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to count files")
	}

	return &ListFilesResponse{
		Files: ToFileResponses(files),
		Meta:  pagination.NewResult(total, &req.Request),
	}, nil
}

func (s *service) List(ctx context.Context, req *ListFilesRequest) (*ListFilesResponse, error) {
	files, err := s.fileRepo.List(ctx, req.Limit(), req.Offset())
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to list files")
	}

	total, err := s.fileRepo.Count(ctx)
	if err != nil {
		return nil, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to count files")
	}

	return &ListFilesResponse{
		Files: ToFileResponses(files),
		Meta:  pagination.NewResult(total, &req.Request),
	}, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID, isAdmin bool) error {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFoundI18n("file.not_found")
		}
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to get file")
	}

	if file.UserID != userID && !isAdmin {
		return apperror.ForbiddenI18n("file.delete_own_only")
	}

	// Begin transaction
	tx, err := s.fileRepo.BeginTx(ctx)
	if err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to begin transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Delete DB record within transaction first
	if err = s.fileRepo.DeleteTx(ctx, tx, id); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to delete file record")
	}

	// Delete from storage
	storageDriver := s.storageManager.Get(file.StorageDriver)
	if err = storageDriver.Delete(ctx, file.StoragePath); err != nil {
		// Rollback DB deletion if storage delete fails
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to delete file from storage")
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternalError, "failed to commit transaction")
	}

	return nil
}
