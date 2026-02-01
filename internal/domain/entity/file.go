package entity

import (
	"time"

	"github.com/google/uuid"
)

type StorageDriver string

const (
	StorageLocal StorageDriver = "local"
	StorageS3    StorageDriver = "s3"
	StorageR2    StorageDriver = "r2"
)

func (s StorageDriver) String() string {
	return string(s)
}

func (s StorageDriver) IsValid() bool {
	switch s {
	case StorageLocal, StorageS3, StorageR2:
		return true
	}
	return false
}

func ParseStorageDriver(s string) StorageDriver {
	switch s {
	case "s3":
		return StorageS3
	case "r2":
		return StorageR2
	default:
		return StorageLocal
	}
}

type File struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Filename      string
	OriginalName  string
	MimeType      string
	Size          int64
	StorageDriver StorageDriver
	StoragePath   string
	URL           string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
