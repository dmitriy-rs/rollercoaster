package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// BenchmarkFileOperations_WithoutCache benchmarks file operations without caching
func BenchmarkFileOperations_WithoutCache(b *testing.B) {
	tmpDir := b.TempDir()

	// Create test files
	testFiles := []string{"package.json", "yarn.lock", "pnpm-lock.yaml", "Taskfile.yml"}
	for _, filename := range testFiles {
		content := `{"name": "test", "version": "1.0.0"}`
		if filepath.Ext(filename) == ".yml" || filepath.Ext(filename) == ".yaml" {
			content = "version: '3'\ntasks:\n  test:\n    desc: 'Test task'"
		}
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(b, err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate typical operations without cache
		for _, filename := range testFiles {
			fullPath := filepath.Join(tmpDir, filename)

			// Check if file exists
			_, err := os.Stat(fullPath)
			if err == nil {
				// Read file content
				_, err := os.ReadFile(fullPath)
				require.NoError(b, err)
			}
		}

		// Read directory
		_, err := os.ReadDir(tmpDir)
		require.NoError(b, err)
	}
}

// BenchmarkFileOperations_WithCache benchmarks file operations with caching
func BenchmarkFileOperations_WithCache(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFSCache(5 * time.Second)

	// Create test files
	testFiles := []string{"package.json", "yarn.lock", "pnpm-lock.yaml", "Taskfile.yml"}
	for _, filename := range testFiles {
		content := `{"name": "test", "version": "1.0.0"}`
		if filepath.Ext(filename) == ".yml" || filepath.Ext(filename) == ".yaml" {
			content = "version: '3'\ntasks:\n  test:\n    desc: 'Test task'"
		}
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(b, err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate typical operations with cache
		for _, filename := range testFiles {
			fullPath := filepath.Join(tmpDir, filename)

			// Check if file exists (cached)
			if cache.FileExists(fullPath) {
				// Read file content (cached)
				_, err := cache.ReadFile(fullPath)
				require.NoError(b, err)
			}
		}

		// Read directory (cached)
		_, err := cache.ReadDir(tmpDir)
		require.NoError(b, err)
	}
}

// BenchmarkBatchFileOperations_WithoutCache benchmarks individual file checks
func BenchmarkBatchFileOperations_WithoutCache(b *testing.B) {
	tmpDir := b.TempDir()

	// Create some test files
	existingFiles := []string{"package.json", "yarn.lock"}
	for _, filename := range existingFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644)
		require.NoError(b, err)
	}

	searchFiles := []string{"package.json", "yarn.lock", "pnpm-lock.yaml", "package-lock.json"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Individual file existence checks
		for _, filename := range searchFiles {
			fullPath := filepath.Join(tmpDir, filename)
			_, err := os.Stat(fullPath)
			_ = err == nil // Just check existence
		}
	}
}

// BenchmarkBatchFileOperations_WithCache benchmarks batch file operations
func BenchmarkBatchFileOperations_WithCache(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFSCache(5 * time.Second)

	// Create some test files
	existingFiles := []string{"package.json", "yarn.lock"}
	for _, filename := range existingFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644)
		require.NoError(b, err)
	}

	searchFiles := []string{"package.json", "yarn.lock", "pnpm-lock.yaml", "package-lock.json"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Batch file existence check
		_ = cache.FindFilesInDirectory(tmpDir, searchFiles)
	}
}

// BenchmarkDirectoryTraversal_WithoutCache benchmarks git directory finding without cache
func BenchmarkDirectoryTraversal_WithoutCache(b *testing.B) {
	tmpDir := b.TempDir()

	// Create nested directory structure
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	err := os.MkdirAll(deepDir, 0755)
	require.NoError(b, err)

	// Create .git directory at root
	gitDir := filepath.Join(tmpDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate finding git directory without cache
		currentDir := deepDir
		for {
			gitPath := filepath.Join(currentDir, ".git")
			info, err := os.Stat(gitPath)
			if err == nil && info.IsDir() {
				break
			}
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir {
				break
			}
			currentDir = parentDir
		}
	}
}

// BenchmarkDirectoryTraversal_WithCache benchmarks git directory finding with cache
func BenchmarkDirectoryTraversal_WithCache(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFSCache(5 * time.Second)

	// Create nested directory structure
	deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	err := os.MkdirAll(deepDir, 0755)
	require.NoError(b, err)

	// Create .git directory at root
	gitDir := filepath.Join(tmpDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Simulate finding git directory with cache
		currentDir := deepDir
		for {
			gitPath := filepath.Join(currentDir, ".git")
			info, err := cache.Stat(gitPath)
			if err == nil && info.IsDir() {
				break
			}
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir {
				break
			}
			currentDir = parentDir
		}
	}
}
