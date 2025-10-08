package scanner_test

import (
	"context"
	"io/fs"
	"testing"

	"github.com/jamesainslie/dot/internal/scanner"
	"github.com/jamesainslie/dot/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFS implements the FS interface for testing scanner logic.
type MockFS struct {
	mock.Mock
}

func (m *MockFS) Stat(ctx context.Context, name string) (domain.FileInfo, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(domain.FileInfo), args.Error(1)
}

func (m *MockFS) ReadDir(ctx context.Context, name string) ([]domain.DirEntry, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DirEntry), args.Error(1)
}

func (m *MockFS) ReadLink(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *MockFS) ReadFile(ctx context.Context, name string) ([]byte, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFS) WriteFile(ctx context.Context, name string, data []byte, perm fs.FileMode) error {
	args := m.Called(ctx, name, data, perm)
	return args.Error(0)
}

func (m *MockFS) Mkdir(ctx context.Context, name string, perm fs.FileMode) error {
	args := m.Called(ctx, name, perm)
	return args.Error(0)
}

func (m *MockFS) MkdirAll(ctx context.Context, name string, perm fs.FileMode) error {
	args := m.Called(ctx, name, perm)
	return args.Error(0)
}

func (m *MockFS) Remove(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFS) RemoveAll(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFS) Symlink(ctx context.Context, oldname, newname string) error {
	args := m.Called(ctx, oldname, newname)
	return args.Error(0)
}

func (m *MockFS) Rename(ctx context.Context, oldname, newname string) error {
	args := m.Called(ctx, oldname, newname)
	return args.Error(0)
}

func (m *MockFS) Exists(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

func (m *MockFS) IsDir(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockFS) IsSymlink(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func TestScanTree_SingleFile(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/file.txt").Unwrap()

	// Mock: path is not a symlink, and is a file (not a directory)
	mockFS.On("IsSymlink", ctx, "/test/file.txt").Return(false, nil)
	mockFS.On("IsDir", ctx, "/test/file.txt").Return(false, nil)

	result := scanner.ScanTree(ctx, mockFS, path)
	require.True(t, result.IsOk())

	node := result.Unwrap()
	assert.Equal(t, path, node.Path)
	assert.Equal(t, domain.NodeFile, node.Type)
	assert.Nil(t, node.Children)

	mockFS.AssertExpectations(t)
}

func TestScanTree_EmptyDirectory(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/dir").Unwrap()

	// Mock: path is not a symlink, is a directory with no children
	mockFS.On("IsSymlink", ctx, "/test/dir").Return(false, nil)
	mockFS.On("IsDir", ctx, "/test/dir").Return(true, nil)
	mockFS.On("ReadDir", ctx, "/test/dir").Return([]domain.DirEntry{}, nil)

	result := scanner.ScanTree(ctx, mockFS, path)
	require.True(t, result.IsOk())

	node := result.Unwrap()
	assert.Equal(t, path, node.Path)
	assert.Equal(t, domain.NodeDir, node.Type)
	assert.Empty(t, node.Children)

	mockFS.AssertExpectations(t)
}

func TestScanTree_Symlink(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/link").Unwrap()

	// Mock: path is a symlink
	mockFS.On("IsSymlink", ctx, "/test/link").Return(true, nil)

	result := scanner.ScanTree(ctx, mockFS, path)
	require.True(t, result.IsOk())

	node := result.Unwrap()
	assert.Equal(t, path, node.Path)
	assert.Equal(t, domain.NodeSymlink, node.Type)
	assert.Nil(t, node.Children)

	mockFS.AssertExpectations(t)
}

func TestScanTree_Error(t *testing.T) {
	ctx := context.Background()
	mockFS := new(MockFS)

	path := domain.NewFilePath("/test/error").Unwrap()

	// Mock: IsSymlink returns an error
	mockFS.On("IsSymlink", ctx, "/test/error").Return(false, assert.AnError)

	result := scanner.ScanTree(ctx, mockFS, path)
	assert.True(t, result.IsErr())

	mockFS.AssertExpectations(t)
}

func TestWalk(t *testing.T) {
	// Build a simple tree: dir -> file1, file2
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1").Unwrap(),
				Type: domain.NodeFile,
			},
			{
				Path: domain.NewFilePath("/test/file2").Unwrap(),
				Type: domain.NodeFile,
			},
		},
	}

	// Collect all visited paths
	var visited []string
	err := scanner.Walk(root, func(n domain.Node) error {
		visited = append(visited, n.Path.String())
		return nil
	})

	require.NoError(t, err)
	assert.Len(t, visited, 3) // root + 2 children
	assert.Contains(t, visited, "/test")
	assert.Contains(t, visited, "/test/file1")
	assert.Contains(t, visited, "/test/file2")
}

func TestWalk_ErrorStopsTraversal(t *testing.T) {
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1").Unwrap(),
				Type: domain.NodeFile,
			},
		},
	}

	// Return error on first visit
	err := scanner.Walk(root, func(n domain.Node) error {
		return assert.AnError
	})

	assert.Error(t, err)
}

func TestCollectFiles(t *testing.T) {
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1.txt").Unwrap(),
				Type: domain.NodeFile,
			},
			{
				Path: domain.NewFilePath("/test/subdir").Unwrap(),
				Type: domain.NodeDir,
				Children: []domain.Node{
					{
						Path: domain.NewFilePath("/test/subdir/file2.txt").Unwrap(),
						Type: domain.NodeFile,
					},
				},
			},
		},
	}

	files := scanner.CollectFiles(root)
	assert.Len(t, files, 2)

	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = f.String()
	}

	assert.Contains(t, paths, "/test/file1.txt")
	assert.Contains(t, paths, "/test/subdir/file2.txt")
}

func TestCountNodes(t *testing.T) {
	root := domain.Node{
		Path: domain.NewFilePath("/test").Unwrap(),
		Type: domain.NodeDir,
		Children: []domain.Node{
			{
				Path: domain.NewFilePath("/test/file1").Unwrap(),
				Type: domain.NodeFile,
			},
			{
				Path: domain.NewFilePath("/test/dir").Unwrap(),
				Type: domain.NodeDir,
				Children: []domain.Node{
					{
						Path: domain.NewFilePath("/test/dir/file2").Unwrap(),
						Type: domain.NodeFile,
					},
				},
			},
		},
	}

	count := scanner.CountNodes(root)
	assert.Equal(t, 4, count) // root + file1 + dir + file2
}

func TestRelativePath(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		target   string
		expected string
		wantErr  bool
	}{
		{
			name:     "same directory",
			base:     "/home/user/.dotfiles",
			target:   "/home/user/.dotfiles/file.txt",
			expected: "file.txt",
			wantErr:  false,
		},
		{
			name:     "nested directory",
			base:     "/home/user/.dotfiles",
			target:   "/home/user/.dotfiles/vim/vimrc",
			expected: "vim/vimrc",
			wantErr:  false,
		},
		{
			name:     "same path",
			base:     "/home/user/.dotfiles",
			target:   "/home/user/.dotfiles",
			expected: ".",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := domain.NewFilePath(tt.base).Unwrap()
			target := domain.NewFilePath(tt.target).Unwrap()

			result := scanner.RelativePath(base, target)

			if tt.wantErr {
				assert.True(t, result.IsErr())
			} else {
				require.True(t, result.IsOk())
				assert.Equal(t, tt.expected, result.Unwrap())
			}
		})
	}
}
