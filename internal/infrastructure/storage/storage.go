package storage

import (
	"context"
	"io"
	"mime/multipart"

	"backend-gin/internal/domain/entity"
)

type UploadResult struct {
	Filename    string
	StoragePath string
	URL         string
	Size        int64
	MimeType    string
}

type Storage interface {
	Upload(ctx context.Context, file *multipart.FileHeader, path string) (*UploadResult, error)
	UploadReader(ctx context.Context, reader io.Reader, filename, path, mimeType string, size int64) (*UploadResult, error)
	Delete(ctx context.Context, path string) error
	GetURL(ctx context.Context, path string) (string, error)
	Driver() entity.StorageDriver
}

type Manager struct {
	drivers map[entity.StorageDriver]Storage
	default_ entity.StorageDriver
}

func NewManager(defaultDriver entity.StorageDriver) *Manager {
	return &Manager{
		drivers:  make(map[entity.StorageDriver]Storage),
		default_: defaultDriver,
	}
}

func (m *Manager) Register(driver Storage) {
	m.drivers[driver.Driver()] = driver
}

func (m *Manager) Get(driver entity.StorageDriver) Storage {
	if s, ok := m.drivers[driver]; ok {
		return s
	}
	return m.drivers[m.default_]
}

func (m *Manager) Default() Storage {
	return m.drivers[m.default_]
}

func (m *Manager) SetDefault(driver entity.StorageDriver) {
	m.default_ = driver
}
