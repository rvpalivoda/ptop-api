package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Service предоставляет методы для загрузки файлов и генерации URL.
type Service struct {
	client *minio.Client
	bucket string
}

// New создаёт новый сервис хранения.
func New(endpoint, accessKey, secretKey, bucket string, useSSL bool) (Storage, error) {
	if endpoint == "" {
		return newMemory(), nil
	}
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Service{client: cli, bucket: bucket}, nil
}

// Upload загружает объект в хранилище.
func (s *Service) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectName, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return objectName, nil
}

// GetURL генерирует временный URL для объекта.
func (s *Service) GetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Storage описывает интерфейс сервиса хранения.
type Storage interface {
	Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) (string, error)
	GetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

var _ Storage = (*Service)(nil)
