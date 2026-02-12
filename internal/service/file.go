package service

import (
	"context"
	"fmt"
	"strings"

	"file_grpc/internal/repository"
)

type FileService struct {
	repo repository.Repository
}

func NewFileService(repo repository.Repository) *FileService {
	return &FileService{repo: repo}
}

// Сохраним файл с проверкой на безопасное написание
func (s *FileService) SaveFile(ctx context.Context, filename string, data []byte) error {
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return fmt.Errorf("invalid filename: %s", filename)
	}
	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}
	return s.repo.Save(ctx, filename, data)
}

// Вернем содержимое файла
func (s *FileService) GetFile(ctx context.Context, filename string) ([]byte, error) {
	return s.repo.Get(ctx, filename)
}

// ListFiles возвращает список файлов с метаданными.
func (s *FileService) ListFiles(ctx context.Context) ([]repository.FileMeta, error) {
	return s.repo.List(ctx)
}

// UpdateAccess обновляет дату последнего доступа.
func (s *FileService) UpdateAccess(ctx context.Context, filename string) error {
	return s.repo.UpdateAccess(ctx, filename)
}
