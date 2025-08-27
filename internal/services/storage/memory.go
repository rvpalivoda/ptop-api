package storage

import (
	"context"
	"io"
	"time"
)

// memoryStorage — простая in-memory реализация Storage для дев-режима.
type memoryStorage struct{}

func newMemory() Storage {
	return &memoryStorage{}
}

func (m *memoryStorage) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) (string, error) {
	// Читаем данные, но никуда их не сохраняем
	if _, err := io.Copy(io.Discard, r); err != nil {
		return "", err
	}
	return objectName, nil
}

func (m *memoryStorage) GetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	// Возвращаем фиктивный URL
	return "https://example.com/" + objectName, nil
}

var _ Storage = (*memoryStorage)(nil)
