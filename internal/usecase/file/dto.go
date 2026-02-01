package file

import (
	"time"

	"github.com/google/uuid"

	"backend-gin/internal/domain/entity"
)

type UploadRequest struct {
	Path string `form:"path"`
}

type FileResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Filename      string    `json:"filename"`
	OriginalName  string    `json:"original_name"`
	MimeType      string    `json:"mime_type"`
	Size          int64     `json:"size"`
	StorageDriver string    `json:"storage_driver"`
	URL           string    `json:"url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ListFilesRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type ListFilesResponse struct {
	Files      []*FileResponse `json:"files"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func ToFileResponse(file *entity.File) *FileResponse {
	return &FileResponse{
		ID:            file.ID,
		UserID:        file.UserID,
		Filename:      file.Filename,
		OriginalName:  file.OriginalName,
		MimeType:      file.MimeType,
		Size:          file.Size,
		StorageDriver: file.StorageDriver.String(),
		URL:           file.URL,
		CreatedAt:     file.CreatedAt,
		UpdatedAt:     file.UpdatedAt,
	}
}

func ToFileResponses(files []*entity.File) []*FileResponse {
	responses := make([]*FileResponse, len(files))
	for i, file := range files {
		responses[i] = ToFileResponse(file)
	}
	return responses
}
