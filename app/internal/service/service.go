// Package service contains the business logic layer of the MinIO service.
// It defines service interfaces and implements use cases by orchestrating
// the S3 repository, applying business rules, and returning results to handlers.
package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"hotel.com/app/internal/models"
	"hotel.com/app/internal/repo"
)

// Service defines all business operations exposed to the handler layer.
type Service interface {
	UploadFile(ctx context.Context, bucket, filename, objectKey string, file io.Reader, size int64, contentType string) (*models.UploadResponse, error)
	DownloadFile(ctx context.Context, bucket, key string) (*models.DownloadResult, error)
}

type minioService struct {
	l  *slog.Logger
	s3 repo.S3Repository
}

// New constructs a Service wired to an S3 repository.
func New(l *slog.Logger, s3 repo.S3Repository) Service {
	return &minioService{
		l:  l,
		s3: s3,
	}
}

// UploadFile validates the input and stores the file in S3.
// When objectKey is provided it is used as the S3 object key (after
// sanitisation), preserving the full prefix path (e.g.
// "hotels/{hotel_id}/{timestamp}-{name}").  When objectKey is empty the
// filename is used instead (legacy / direct-upload behaviour).
func (s *minioService) UploadFile(ctx context.Context, bucket, filename, objectKey string, file io.Reader, size int64, contentType string) (*models.UploadResponse, error) {
	if strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("bucket name must not be empty")
	}

	// Determine the key: prefer the explicit objectKey, fall back to filename.
	rawKey := objectKey
	if strings.TrimSpace(rawKey) == "" {
		rawKey = filename
	}
	if strings.TrimSpace(rawKey) == "" {
		return nil, fmt.Errorf("object key must not be empty")
	}

	// Sanitise: collapse any path-traversal sequences but keep the full
	// prefix path so keys like "hotels/{id}/{ts}-file.jpg" are preserved.
	key := sanitiseKey(rawKey)

	if err := s.s3.EnsureBucket(ctx, bucket); err != nil {
		s.l.Error("ensure bucket failed", "bucket", bucket, "err", err)
		return nil, err
	}

	if err := s.s3.UploadFile(ctx, bucket, key, file, size, contentType); err != nil {
		s.l.Error("upload failed", "bucket", bucket, "key", key, "err", err)
		return nil, err
	}

	s.l.Info("file uploaded", "bucket", bucket, "key", key)
	return &models.UploadResponse{Bucket: bucket, Key: key}, nil
}

// sanitiseKey removes path-traversal components (../) from the key while
// preserving legitimate prefix directories such as "hotels/{id}/".
func sanitiseKey(key string) string {
	parts := strings.Split(key, "/")
	var clean []string
	for _, p := range parts {
		if p == ".." || p == "." || p == "" {
			continue
		}
		clean = append(clean, p)
	}
	return strings.Join(clean, "/")
}

// DownloadFile fetches the object from S3 and returns the streaming result.
func (s *minioService) DownloadFile(ctx context.Context, bucket, key string) (*models.DownloadResult, error) {
	if strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("bucket name must not be empty")
	}
	if strings.TrimSpace(key) == "" {
		return nil, fmt.Errorf("object key must not be empty")
	}

	result, err := s.s3.DownloadFile(ctx, bucket, key)
	if err != nil {
		s.l.Error("download failed", "bucket", bucket, "key", key, "err", err)
		return nil, err
	}

	s.l.Info("file downloaded", "bucket", bucket, "key", key, "size", result.Size)
	return result, nil
}
