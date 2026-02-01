package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"backend-gin/internal/domain/entity"
)

type LocalStorage struct {
	basePath string
	baseURL  string
}

type LocalConfig struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(cfg LocalConfig) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: cfg.BasePath,
		baseURL:  cfg.BaseURL,
	}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, file *multipart.FileHeader, path string) (*UploadResult, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	filename := s.generateFilename(file.Filename)
	fullPath := filepath.Join(path, filename)
	storagePath := filepath.Join(s.basePath, fullPath)

	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return &UploadResult{
		Filename:    filename,
		StoragePath: fullPath,
		URL:         s.baseURL + "/" + fullPath,
		Size:        size,
		MimeType:    file.Header.Get("Content-Type"),
	}, nil
}

func (s *LocalStorage) UploadReader(ctx context.Context, reader io.Reader, filename, path, mimeType string, size int64) (*UploadResult, error) {
	generatedFilename := s.generateFilename(filename)
	fullPath := filepath.Join(path, generatedFilename)
	storagePath := filepath.Join(s.basePath, fullPath)

	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	dst, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return &UploadResult{
		Filename:    generatedFilename,
		StoragePath: fullPath,
		URL:         s.baseURL + "/" + fullPath,
		Size:        written,
		MimeType:    mimeType,
	}, nil
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	return s.baseURL + "/" + path, nil
}

func (s *LocalStorage) Driver() entity.StorageDriver {
	return entity.StorageLocal
}

func (s *LocalStorage) generateFilename(original string) string {
	ext := filepath.Ext(original)
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
}
