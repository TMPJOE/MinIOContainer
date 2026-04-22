// Package repo implements the data access layer of the MinIO service.
// It handles all S3 operations via the MinIO Go SDK.
package repo

import (
	"context"
	"io"

	"hotel.com/app/internal/models"
)

// S3Repository defines the object-storage operations required by the service layer.
type S3Repository interface {
	// UploadFile stores a file in the given bucket under objectName.
	// contentType should be the MIME type (e.g. "image/png").
	UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, size int64, contentType string) error

	// DownloadFile retrieves a file from the given bucket by its object key.
	DownloadFile(ctx context.Context, bucketName, objectName string) (*models.DownloadResult, error)

	// EnsureBucket creates the bucket if it does not already exist.
	EnsureBucket(ctx context.Context, bucketName string) error
}
