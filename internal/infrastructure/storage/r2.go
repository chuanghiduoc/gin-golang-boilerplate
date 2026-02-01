package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"backend-gin/internal/domain/entity"
)

type R2Storage struct {
	client     *s3.Client
	bucket     string
	baseURL    string
	pathPrefix string
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	PathPrefix      string
}

func NewR2Storage(ctx context.Context, cfg R2Config) (*R2Storage, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &R2Storage{
		client:     client,
		bucket:     cfg.Bucket,
		baseURL:    cfg.BaseURL,
		pathPrefix: cfg.PathPrefix,
	}, nil
}

func (s *R2Storage) Upload(ctx context.Context, file *multipart.FileHeader, path string) (*UploadResult, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	filename := s.generateFilename(file.Filename)
	key := s.buildKey(path, filename)
	contentType := file.Header.Get("Content-Type")

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          src,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(file.Size),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	return &UploadResult{
		Filename:    filename,
		StoragePath: key,
		URL:         s.baseURL + "/" + key,
		Size:        file.Size,
		MimeType:    contentType,
	}, nil
}

func (s *R2Storage) UploadReader(ctx context.Context, reader io.Reader, filename, path, mimeType string, size int64) (*UploadResult, error) {
	generatedFilename := s.generateFilename(filename)
	key := s.buildKey(path, generatedFilename)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(mimeType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to R2: %w", err)
	}

	return &UploadResult{
		Filename:    generatedFilename,
		StoragePath: key,
		URL:         s.baseURL + "/" + key,
		Size:        size,
		MimeType:    mimeType,
	}, nil
}

func (s *R2Storage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}
	return nil
}

func (s *R2Storage) GetURL(ctx context.Context, path string) (string, error) {
	return s.baseURL + "/" + path, nil
}

func (s *R2Storage) Driver() entity.StorageDriver {
	return entity.StorageR2
}

func (s *R2Storage) buildKey(path, filename string) string {
	if s.pathPrefix != "" {
		return filepath.Join(s.pathPrefix, path, filename)
	}
	return filepath.Join(path, filename)
}

func (s *R2Storage) generateFilename(original string) string {
	ext := filepath.Ext(original)
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
}
