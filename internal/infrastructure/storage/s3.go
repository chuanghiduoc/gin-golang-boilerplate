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

type S3Storage struct {
	client     *s3.Client
	bucket     string
	baseURL    string
	pathPrefix string
}

type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	BaseURL         string
	PathPrefix      string
	Endpoint        string
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	var client *s3.Client
	if cfg.Endpoint != "" {
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
	}

	return &S3Storage{
		client:     client,
		bucket:     cfg.Bucket,
		baseURL:    baseURL,
		pathPrefix: cfg.PathPrefix,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, file *multipart.FileHeader, path string) (*UploadResult, error) {
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
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &UploadResult{
		Filename:    filename,
		StoragePath: key,
		URL:         s.baseURL + "/" + key,
		Size:        file.Size,
		MimeType:    contentType,
	}, nil
}

func (s *S3Storage) UploadReader(ctx context.Context, reader io.Reader, filename, path, mimeType string, size int64) (*UploadResult, error) {
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
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &UploadResult{
		Filename:    generatedFilename,
		StoragePath: key,
		URL:         s.baseURL + "/" + key,
		Size:        size,
		MimeType:    mimeType,
	}, nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}
	return nil
}

func (s *S3Storage) GetURL(ctx context.Context, path string) (string, error) {
	return s.baseURL + "/" + path, nil
}

func (s *S3Storage) Driver() entity.StorageDriver {
	return entity.StorageS3
}

func (s *S3Storage) buildKey(path, filename string) string {
	if s.pathPrefix != "" {
		return filepath.Join(s.pathPrefix, path, filename)
	}
	return filepath.Join(path, filename)
}

func (s *S3Storage) generateFilename(original string) string {
	ext := filepath.Ext(original)
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), uuid.New().String()[:8], ext)
}
