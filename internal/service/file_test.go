package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"file_grpc/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------
// Mock Repository
// ---------------------------------------------------------------------
type mockRepo struct {
	saveFunc         func(ctx context.Context, filename string, data []byte) error
	getFunc          func(ctx context.Context, filename string) ([]byte, error)
	listFunc         func(ctx context.Context) ([]repository.FileMeta, error)
	updateAccessFunc func(ctx context.Context, filename string) error
}

func (m *mockRepo) Save(ctx context.Context, filename string, data []byte) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, filename, data)
	}
	return nil
}

func (m *mockRepo) Get(ctx context.Context, filename string) ([]byte, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, filename)
	}
	return nil, nil
}

func (m *mockRepo) List(ctx context.Context) ([]repository.FileMeta, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockRepo) UpdateAccess(ctx context.Context, filename string) error {
	if m.updateAccessFunc != nil {
		return m.updateAccessFunc(ctx, filename)
	}
	return nil
}

// ---------------------------------------------------------------------
// SaveFile
// ---------------------------------------------------------------------
func TestFileService_SaveFile(t *testing.T) {
	ctx := context.Background()

	t.Run("successful save", func(t *testing.T) {
		var capturedFilename string
		var capturedData []byte

		mock := &mockRepo{
			saveFunc: func(ctx context.Context, filename string, data []byte) error {
				capturedFilename = filename
				capturedData = data
				return nil
			},
		}
		svc := NewFileService(mock)

		filename := "valid.txt"
		data := []byte("hello")
		err := svc.SaveFile(ctx, filename, data)
		require.NoError(t, err)
		assert.Equal(t, filename, capturedFilename)
		assert.Equal(t, data, capturedData)
	})

	t.Run("invalid filename with ..", func(t *testing.T) {
		mock := &mockRepo{}
		svc := NewFileService(mock)

		err := svc.SaveFile(ctx, "../evil.txt", []byte("bad"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid filename")
	})

	t.Run("invalid filename with /", func(t *testing.T) {
		svc := NewFileService(&mockRepo{})
		err := svc.SaveFile(ctx, "sub/dir/file.txt", []byte("bad"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid filename")
	})

	t.Run("invalid filename with \\", func(t *testing.T) {
		svc := NewFileService(&mockRepo{})
		err := svc.SaveFile(ctx, "sub\\dir\\file.txt", []byte("bad"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid filename")
	})

	t.Run("empty data", func(t *testing.T) {
		svc := NewFileService(&mockRepo{})
		err := svc.SaveFile(ctx, "empty.txt", []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty file")
	})

	t.Run("repository error propagation", func(t *testing.T) {
		expectedErr := errors.New("disk full")
		mock := &mockRepo{
			saveFunc: func(ctx context.Context, filename string, data []byte) error {
				return expectedErr
			},
		}
		svc := NewFileService(mock)
		err := svc.SaveFile(ctx, "valid.txt", []byte("data"))
		assert.ErrorIs(t, err, expectedErr)
	})
}

// ---------------------------------------------------------------------
// GetFile
// ---------------------------------------------------------------------
func TestFileService_GetFile(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		expectedData := []byte("content")
		mock := &mockRepo{
			getFunc: func(ctx context.Context, filename string) ([]byte, error) {
				return expectedData, nil
			},
		}
		svc := NewFileService(mock)

		data, err := svc.GetFile(ctx, "any.txt")
		require.NoError(t, err)
		assert.Equal(t, expectedData, data)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("file not found")
		mock := &mockRepo{
			getFunc: func(ctx context.Context, filename string) ([]byte, error) {
				return nil, expectedErr
			},
		}
		svc := NewFileService(mock)

		_, err := svc.GetFile(ctx, "missing.txt")
		assert.ErrorIs(t, err, expectedErr)
	})
}

// ---------------------------------------------------------------------
// ListFiles
// ---------------------------------------------------------------------
func TestFileService_ListFiles(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		now := time.Now()
		expectedMetas := []repository.FileMeta{
			{Filename: "a.txt", Size: 10, CreatedAt: now, UpdatedAt: now},
			{Filename: "b.jpg", Size: 20, CreatedAt: now, UpdatedAt: now},
		}
		mock := &mockRepo{
			listFunc: func(ctx context.Context) ([]repository.FileMeta, error) {
				return expectedMetas, nil
			},
		}
		svc := NewFileService(mock)

		metas, err := svc.ListFiles(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedMetas, metas)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("can't list")
		mock := &mockRepo{
			listFunc: func(ctx context.Context) ([]repository.FileMeta, error) {
				return nil, expectedErr
			},
		}
		svc := NewFileService(mock)

		_, err := svc.ListFiles(ctx)
		assert.ErrorIs(t, err, expectedErr)
	})
}

// ---------------------------------------------------------------------
// UpdateAccess
// ---------------------------------------------------------------------
func TestFileService_UpdateAccess(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		var calledFilename string
		mock := &mockRepo{
			updateAccessFunc: func(ctx context.Context, filename string) error {
				calledFilename = filename
				return nil
			},
		}
		svc := NewFileService(mock)

		err := svc.UpdateAccess(ctx, "test.txt")
		require.NoError(t, err)
		assert.Equal(t, "test.txt", calledFilename)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("update failed")
		mock := &mockRepo{
			updateAccessFunc: func(ctx context.Context, filename string) error {
				return expectedErr
			},
		}
		svc := NewFileService(mock)

		err := svc.UpdateAccess(ctx, "test.txt")
		assert.ErrorIs(t, err, expectedErr)
	})
}
