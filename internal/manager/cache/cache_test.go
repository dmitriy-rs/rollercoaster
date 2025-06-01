package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFSCache_Stat(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	// First call should hit the file system
	info1, err1 := cache.Stat(tmpFile)
	require.NoError(t, err1)

	// Second call should use cache
	info2, err2 := cache.Stat(tmpFile)
	require.NoError(t, err2)

	// Verify same result
	assert.Equal(t, info1.Name(), info2.Name())
	assert.Equal(t, info1.Size(), info2.Size())

	// Test cache expiration
	time.Sleep(150 * time.Millisecond)
	info3, err3 := cache.Stat(tmpFile)
	require.NoError(t, err3)

	assert.Equal(t, info1.Name(), info3.Name())
}

func TestFSCache_FileExists(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)

	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	nonExistingFile := filepath.Join(tmpDir, "notexists.txt")

	err := os.WriteFile(existingFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Test existing file
	assert.True(t, cache.FileExists(existingFile), "FileExists should return true for existing file")

	// Test non-existing file
	assert.False(t, cache.FileExists(nonExistingFile), "FileExists should return false for non-existing file")

	// Test cache is used (second call should be faster)
	start := time.Now()
	cache.FileExists(existingFile)
	duration := time.Since(start)

	assert.Less(t, duration, 1*time.Millisecond, "Second FileExists call should be much faster (cached)")
}

func TestFSCache_ReadDir(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)

	tmpDir := t.TempDir()

	// Create some test files
	testFiles := []string{"file1.txt", "file2.js", "file3.json"}
	for _, filename := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644)
		require.NoError(t, err)
	}

	// First call
	entries1, err1 := cache.ReadDir(tmpDir)
	require.NoError(t, err1)

	// Second call (should be cached)
	entries2, err2 := cache.ReadDir(tmpDir)
	require.NoError(t, err2)

	// Verify results are the same
	assert.Equal(t, len(entries1), len(entries2), "Cached ReadDir result has different length")

	for i, entry := range entries1 {
		assert.Equal(t, entry.Name(), entries2[i].Name(), "Cached ReadDir result differs from original")
	}
}

func TestFSCache_ReadFile(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	expectedContent := []byte("Hello, World!")

	err := os.WriteFile(tmpFile, expectedContent, 0644)
	require.NoError(t, err)

	// First call
	content1, err1 := cache.ReadFile(tmpFile)
	require.NoError(t, err1)

	// Second call (should be cached)
	content2, err2 := cache.ReadFile(tmpFile)
	require.NoError(t, err2)

	// Verify content
	assert.Equal(t, expectedContent, content1, "First ReadFile returned incorrect content")
	assert.Equal(t, expectedContent, content2, "Cached ReadFile returned incorrect content")
	assert.Equal(t, content1, content2, "Cached content differs from original")
}

func TestFSCache_FindFilesInDirectory(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)

	tmpDir := t.TempDir()

	// Create some test files
	existingFiles := []string{"package.json", "yarn.lock", "README.md"}
	for _, filename := range existingFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644)
		require.NoError(t, err)
	}

	searchFiles := []string{"package.json", "yarn.lock", "pnpm-lock.yaml", "package-lock.json"}

	result := cache.FindFilesInDirectory(tmpDir, searchFiles)

	// Verify results
	expected := map[string]bool{
		"package.json":      true,
		"yarn.lock":         true,
		"pnpm-lock.yaml":    false,
		"package-lock.json": false,
	}

	for filename, expectedExists := range expected {
		assert.Equal(t, expectedExists, result[filename],
			"FindFilesInDirectory: expected %s to be %v, got %v", filename, expectedExists, result[filename])
	}
}

func TestFSCache_FindFilesInDirectoryWithContent(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)

	tmpDir := t.TempDir()

	// Create test files with different content
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
	}

	for filename, content := range files {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	searchFiles := []string{"file1.txt", "file2.txt", "file3.txt"}

	result := cache.FindFilesInDirectoryWithContent(tmpDir, searchFiles)

	// Verify file1.txt
	assert.True(t, result["file1.txt"].Exists, "file1.txt should exist")
	assert.Equal(t, "content1", string(result["file1.txt"].Content), "file1.txt content mismatch")

	// Verify file2.txt
	assert.True(t, result["file2.txt"].Exists, "file2.txt should exist")
	assert.Equal(t, "content2", string(result["file2.txt"].Content), "file2.txt content mismatch")

	// Verify file3.txt (non-existing)
	assert.False(t, result["file3.txt"].Exists, "file3.txt should not exist")
}

func TestFSCache_TTLExpiration(t *testing.T) {
	cache := NewFSCache(50 * time.Millisecond)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	err := os.WriteFile(tmpFile, []byte("initial"), 0644)
	require.NoError(t, err)

	// Read file (should be cached)
	content1, err := cache.ReadFile(tmpFile)
	require.NoError(t, err)

	// Modify file
	err = os.WriteFile(tmpFile, []byte("modified"), 0644)
	require.NoError(t, err)

	// Read again immediately (should still return cached content)
	content2, err := cache.ReadFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, content1, content2, "Cache should return same content before TTL expiration")

	// Wait for TTL expiration
	time.Sleep(100 * time.Millisecond)

	// Read again (should return new content)
	content3, err := cache.ReadFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, "modified", string(content3), "Cache should return new content after TTL expiration")
}

func TestFSCache_ClearExpired(t *testing.T) {
	cache := NewFSCache(50 * time.Millisecond)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Add some entries to cache
	cache.Stat(tmpFile)
	cache.ReadDir(tmpDir)
	cache.ReadFile(tmpFile)

	// Check initial stats
	stats1 := cache.GetStats()
	assert.Greater(t, stats1.StatCacheSize, 0, "StatCacheSize should be greater than 0")
	assert.Greater(t, stats1.DirCacheSize, 0, "DirCacheSize should be greater than 0")
	assert.Greater(t, stats1.ContentCacheSize, 0, "ContentCacheSize should be greater than 0")

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Clear expired entries
	cache.ClearExpired()

	// Check stats after clearing
	stats2 := cache.GetStats()
	assert.Equal(t, 0, stats2.StatCacheSize, "StatCacheSize should be 0 after clearing expired entries")
	assert.Equal(t, 0, stats2.DirCacheSize, "DirCacheSize should be 0 after clearing expired entries")
	assert.Equal(t, 0, stats2.ContentCacheSize, "ContentCacheSize should be 0 after clearing expired entries")
}

func TestFSCache_Clear(t *testing.T) {
	cache := NewFSCache(1 * time.Second)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Add entries to cache
	cache.Stat(tmpFile)
	cache.ReadDir(tmpDir)
	cache.ReadFile(tmpFile)

	// Verify cache has entries
	stats1 := cache.GetStats()
	assert.Greater(t, stats1.StatCacheSize, 0, "StatCacheSize should be greater than 0")
	assert.Greater(t, stats1.DirCacheSize, 0, "DirCacheSize should be greater than 0")
	assert.Greater(t, stats1.ContentCacheSize, 0, "ContentCacheSize should be greater than 0")

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	stats2 := cache.GetStats()
	assert.Equal(t, 0, stats2.StatCacheSize, "StatCacheSize should be 0 after Clear()")
	assert.Equal(t, 0, stats2.DirCacheSize, "DirCacheSize should be 0 after Clear()")
	assert.Equal(t, 0, stats2.ContentCacheSize, "ContentCacheSize should be 0 after Clear()")
}
