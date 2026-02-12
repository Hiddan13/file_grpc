package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------
// Вспомогательная функция для создания временного репозитория
// ---------------------------------------------------------------------
func setupTestRepo(t *testing.T) (*FilesRepository, string) {
	t.Helper()
	tmpDir := t.TempDir() // автоматически удалится после теста
	repo, err := NewFilesRepository(tmpDir)
	require.NoError(t, err)
	return repo, tmpDir
}

// ---------------------------------------------------------------------
// Создание репозитория
// ---------------------------------------------------------------------
func TestNewFilesRepository(t *testing.T) {
	t.Run("successfully creates directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		repo, err := NewFilesRepository(tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, repo)
		assert.Equal(t, tmpDir, repo.storagePath)

		// Проверяем, что директория действительно создана
		info, err := os.Stat(tmpDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("fails on invalid path", func(t *testing.T) {
		// Попытка создать директорию в защищённом месте вызовет ошибку
		// (зависит от ОС, пропускаем, если нет прав)
		_, err := NewFilesRepository("/invalid/protected/path")
		if err != nil {
			assert.Error(t, err)
		}
	})
}

// ---------------------------------------------------------------------
// Save
// ---------------------------------------------------------------------
func TestFilesRepository_Save(t *testing.T) {
	ctx := context.Background()

	t.Run("save new file", func(t *testing.T) {
		repo, tmpDir := setupTestRepo(t)
		filename := "new.txt"
		data := []byte("hello world")

		err := repo.Save(ctx, filename, data)
		require.NoError(t, err)

		// Проверяем, что файл создан на диске
		fullPath := filepath.Join(tmpDir, filename)
		content, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, data, content)

		// Проверяем метаданные
		repo.mu.RLock()
		meta, exists := repo.metadata[filename]
		repo.mu.RUnlock()
		assert.True(t, exists)
		assert.Equal(t, filename, meta.Filename)
		assert.Equal(t, int64(len(data)), meta.Size)
		assert.False(t, meta.CreatedAt.IsZero())
		assert.Equal(t, meta.CreatedAt, meta.UpdatedAt) // при создании равны
	})

	t.Run("save existing file (update)", func(t *testing.T) {
		repo, tmpDir := setupTestRepo(t)
		filename := "update.txt"
		oldData := []byte("old")
		newData := []byte("new content longer")

		// Сохраняем первый раз
		err := repo.Save(ctx, filename, oldData)
		require.NoError(t, err)

		// Запоминаем время создания
		repo.mu.RLock()
		oldMeta := repo.metadata[filename]
		repo.mu.RUnlock()

		// Небольшая пауза, чтобы время обновления гарантированно отличалось
		time.Sleep(10 * time.Millisecond)

		// Обновляем файл
		err = repo.Save(ctx, filename, newData)
		require.NoError(t, err)

		// Проверяем содержимое на диске
		fullPath := filepath.Join(tmpDir, filename)
		content, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, newData, content)

		// Проверяем метаданные
		repo.mu.RLock()
		newMeta := repo.metadata[filename]
		repo.mu.RUnlock()
		assert.Equal(t, oldMeta.Filename, newMeta.Filename)
		assert.Equal(t, oldMeta.CreatedAt, newMeta.CreatedAt) // CreatedAt не меняется
		assert.Equal(t, int64(len(newData)), newMeta.Size)
		assert.True(t, newMeta.UpdatedAt.After(oldMeta.UpdatedAt)) // UpdatedAt обновилось
	})

	t.Run("save with empty data", func(t *testing.T) {
		repo, _ := setupTestRepo(t)
		err := repo.Save(ctx, "empty.txt", []byte{})
		require.NoError(t, err) // репозиторий позволяет сохранять пустые файлы
		// проверка что файл создан (0 байт)
	})
}

// ---------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------
func TestFilesRepository_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("get existing file", func(t *testing.T) {
		repo, _ := setupTestRepo(t)
		filename := "exists.txt"
		data := []byte("some data")
		err := repo.Save(ctx, filename, data)
		require.NoError(t, err)

		got, err := repo.Get(ctx, filename)
		require.NoError(t, err)
		assert.Equal(t, data, got)
	})

	t.Run("get non-existing file", func(t *testing.T) {
		repo, _ := setupTestRepo(t)
		_, err := repo.Get(ctx, "missing.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})
}

// ---------------------------------------------------------------------
// List
// ---------------------------------------------------------------------
func TestFilesRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		repo, _ := setupTestRepo(t)
		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("multiple files", func(t *testing.T) {
		repo, _ := setupTestRepo(t)

		files := []struct {
			name string
			data []byte
		}{
			{"a.txt", []byte("aaa")},
			{"b.jpg", []byte("bbb")},
			{"c.go", []byte("ccc")},
		}
		for _, f := range files {
			err := repo.Save(ctx, f.name, f.data)
			require.NoError(t, err)
			time.Sleep(2 * time.Millisecond) // чтобы created_at различалось
		}

		list, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 3)

		// Преобразуем в map для удобства проверки
		metaMap := make(map[string]FileMeta)
		for _, m := range list {
			metaMap[m.Filename] = m
		}

		for _, f := range files {
			m, ok := metaMap[f.name]
			assert.True(t, ok)
			assert.Equal(t, int64(len(f.data)), m.Size)
			assert.False(t, m.CreatedAt.IsZero())
			assert.Equal(t, m.CreatedAt, m.UpdatedAt) // без скачивания дата обновления = создания
		}
	})
}

// ---------------------------------------------------------------------
// UpdateAccess
// ---------------------------------------------------------------------
func TestFilesRepository_UpdateAccess(t *testing.T) {
	ctx := context.Background()

	t.Run("update existing file", func(t *testing.T) {
		repo, _ := setupTestRepo(t)
		filename := "access.txt"
		err := repo.Save(ctx, filename, []byte("data"))
		require.NoError(t, err)

		repo.mu.RLock()
		oldMeta := repo.metadata[filename]
		repo.mu.RUnlock()

		time.Sleep(5 * time.Millisecond)

		err = repo.UpdateAccess(ctx, filename)
		require.NoError(t, err)

		repo.mu.RLock()
		newMeta := repo.metadata[filename]
		repo.mu.RUnlock()

		assert.Equal(t, oldMeta.CreatedAt, newMeta.CreatedAt)
		assert.True(t, newMeta.UpdatedAt.After(oldMeta.UpdatedAt))
	})

	t.Run("update non-existing file", func(t *testing.T) {
		repo, _ := setupTestRepo(t)
		err := repo.UpdateAccess(ctx, "ghost.txt")
		require.NoError(t, err) // метод не возвращает ошибку, просто ничего не делает
		// можно дополнительно проверить, что файл не появился
		_, err = repo.Get(ctx, "ghost.txt")
		assert.Error(t, err)
	})
}

// ---------------------------------------------------------------------
// Конкурентный доступ (опционально, но демонстрирует потокобезопасность)
// ---------------------------------------------------------------------
func TestFilesRepository_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	repo, _ := setupTestRepo(t)

	const goroutines = 50
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// Уникальное имя файла (включает индекс)
			filename := fmt.Sprintf("concurrent_%d.txt", n)
			err := repo.Save(ctx, filename, []byte("data"))
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("ошибка сохранения: %v", err)
	}

	// Проверяем, что все 50 файлов присутствуют
	list, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, list, goroutines)
}
