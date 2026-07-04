// Package storage is the object-storage seam (MinIO / S3-compatible).
//
// The avatar upload streams bytes through the server via Put. To scale out,
// swap in a presigned-URL flow (the client PUTs directly to the bucket using a
// short-lived signed URL) behind this same interface: only the transport
// changes, not the domain. New upload types (documents, exports, ...) reuse Put
// with a different key prefix, so the interface does not grow per type.
package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/isyll/go-grpc-starter/pkg/config"
)

type Storage interface {
	// Put stores an object of the given size and content type under key.
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	// PublicURL returns a retrieval URL for a stored object.
	PublicURL(key string) string
}

type minioStore struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

// NewMinIO connects to the object store and ensures the bucket exists.
func NewMinIO(ctx context.Context, cfg *config.StorageConfig) (Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: connect: %w", err)
	}

	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("storage: create bucket: %w", err)
		}
	}

	return &minioStore{
		client:        client,
		bucket:        cfg.Bucket,
		publicBaseURL: cfg.PublicBaseURL,
	}, nil
}

func (s *minioStore) Put(
	ctx context.Context, key string, r io.Reader, size int64, contentType string,
) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size,
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("storage: put %s: %w", key, err)
	}
	return nil
}

func (s *minioStore) PublicURL(key string) string {
	base := s.publicBaseURL
	if base == "" {
		return key
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(key, "/")
}
