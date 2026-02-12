package repository

import (
	"context"
	"time"
)

type FileMeta struct {
	Filename  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Size      int64
}

type Repository interface {
	// Сохраняет файл на диск + метаданные
	Save(ctx context.Context, filename string, data []byte) error
	// Вернёт содержимое файла
	Get(ctx context.Context, filename string) ([]byte, error)
	// Вернет список всех файлов с метаданными
	List(ctx context.Context) ([]FileMeta, error)
	// Обновляем дату последнего доступа
	UpdateAccess(ctx context.Context, filename string) error
}
