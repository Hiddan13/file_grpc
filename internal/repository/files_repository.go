package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FilesRepository struct {
	storagePath string
	mu          sync.RWMutex
	metadata    map[string]FileMeta
}

func NewFilesRepository(storagePath string) (*FilesRepository, error) {
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo dir: %w", err)
	}
	return &FilesRepository{
		storagePath: storagePath,
		metadata:    make(map[string]FileMeta),
	}, nil
}

func (r *FilesRepository) Save(ctx context.Context, filename string, data []byte) error {
	fullPath := filepath.Join(r.storagePath, filename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if meta, exists := r.metadata[filename]; exists {
		meta.UpdatedAt = now
		meta.Size = int64(len(data))
		r.metadata[filename] = meta
	} else {
		r.metadata[filename] = FileMeta{
			Filename:  filename,
			CreatedAt: now,
			UpdatedAt: now,
			Size:      int64(len(data)),
		}
	}
	return nil
}

func (r *FilesRepository) Get(ctx context.Context, filename string) ([]byte, error) {
	fullPath := filepath.Join(r.storagePath, filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filename)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

func (r *FilesRepository) List(ctx context.Context) ([]FileMeta, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]FileMeta, 0, len(r.metadata))
	for _, meta := range r.metadata {
		list = append(list, meta)
	}
	return list, nil
}

func (r *FilesRepository) UpdateAccess(ctx context.Context, filename string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if meta, exists := r.metadata[filename]; exists {
		meta.UpdatedAt = time.Now()
		r.metadata[filename] = meta
	}
	return nil
}
