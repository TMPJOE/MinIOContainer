package models

import (
	"github.com/minio/minio-go/v7"
)

// DownloadResult holds the streaming body and metadata for a downloaded object.
type DownloadResult struct {
	Object      *minio.Object
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	FileName    string `json:"file_name"`
}

// UploadResponse is returned after a successful file upload.
type UploadResponse struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}
