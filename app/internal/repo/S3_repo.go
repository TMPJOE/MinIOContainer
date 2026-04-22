package repo

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"hotel.com/app/internal/models"
)

// s3Repo implements S3Repository using the MinIO Go SDK.
type s3Repo struct {
	client *minio.Client
}

// NewS3Repo creates a MinIO client and returns an S3Repository.
func NewS3Repo(endpoint, accessKey, secretKey string, useSSL bool) (S3Repository, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}
	return &s3Repo{client: client}, nil
}

// EnsureBucket creates the bucket if it does not already exist.
func (r *s3Repo) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := r.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket existence: %w", err)
	}
	if exists {
		return nil
	}
	if err := r.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket %q: %w", bucketName, err)
	}
	return nil
}

// UploadFile stores the reader content in MinIO under bucketName/objectName.
// Pass size = -1 if the total size is unknown (chunked streaming).
func (r *s3Repo) UploadFile(ctx context.Context, bucketName, objectName string, file io.Reader, size int64, contentType string) error {
	_, err := r.client.PutObject(ctx, bucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("upload object %q to bucket %q: %w", objectName, bucketName, err)
	}
	return nil
}

// DownloadFile retrieves an object from MinIO and returns a DownloadResult
// containing the streaming body and metadata.
func (r *s3Repo) DownloadFile(ctx context.Context, bucketName, objectName string) (*models.DownloadResult, error) {
	obj, err := r.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get object %q from bucket %q: %w", objectName, bucketName, err)
	}

	info, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, fmt.Errorf("stat object %q: %w", objectName, err)
	}

	return &models.DownloadResult{
		Object:      obj,
		ContentType: info.ContentType,
		Size:        info.Size,
		FileName:    info.Key,
	}, nil
}
