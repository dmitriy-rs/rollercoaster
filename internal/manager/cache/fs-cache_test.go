package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFSCache_Stat(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)
	
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// First call should hit the file system
	info1, err1 := cache.Stat(tmpFile)
	if err1 != nil {
		t.Fatal(err1)
	}

	// Second call should use cache
	info2, err2 := cache.Stat(tmpFile)
	if err2 != nil {
		t.Fatal(err2)
	}

	// Verify same result
	if info1.Name() != info2.Name() || info1.Size() != info2.Size() {
		t.Error("Cached stat result differs from original")
	}

	// Test cache expiration
	time.Sleep(150 * time.Millisecond)
	info3, err3 := cache.Stat(tmpFile)
	if err3 != nil {
		t.Fatal(err3)
	}

	if info3.Name() != info1.Name() {
		t.Error("Stat result after cache expiration differs")
	}
}

func TestFSCache_FileExists(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)
	
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "exists.txt")
	nonExistingFile := filepath.Join(tmpDir, "notexists.txt")
	
	err := os.WriteFile(existingFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test existing file
	if !cache.FileExists(existingFile) {
		t.Error("FileExists should return true for existing file")
	}

	// Test non-existing file
	if cache.FileExists(nonExistingFile) {
		t.Error("FileExists should return false for non-existing file")
	}

	// Test cache is used (second call should be faster)
	start := time.Now()
	cache.FileExists(existingFile)
	duration := time.Since(start)
	
	if duration > 1*time.Millisecond {
		t.Error("Second FileExists call should be much faster (cached)")
	}
}

func TestFSCache_ReadDir(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)
	
	tmpDir := t.TempDir()
	
	// Create some test files
	testFiles := []string{"file1.txt", "file2.js", "file3.json"}
	for _, filename := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// First call
	entries1, err1 := cache.ReadDir(tmpDir)
	if err1 != nil {
		t.Fatal(err1)
	}

	// Second call (should be cached)
	entries2, err2 := cache.ReadDir(tmpDir)
	if err2 != nil {
		t.Fatal(err2)
	}

	// Verify results are the same
	if len(entries1) != len(entries2) {
		t.Error("Cached ReadDir result has different length")
	}

	for i, entry := range entries1 {
		if entry.Name() != entries2[i].Name() {
			t.Error("Cached ReadDir result differs from original")
		}
	}
}

func TestFSCache_ReadFile(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)
	
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	expectedContent := []byte("Hello, World!")
	
	err := os.WriteFile(tmpFile, expectedContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// First call
	content1, err1 := cache.ReadFile(tmpFile)
	if err1 != nil {
		t.Fatal(err1)
	}

	// Second call (should be cached)
	content2, err2 := cache.ReadFile(tmpFile)
	if err2 != nil {
		t.Fatal(err2)
	}

	// Verify content
	if string(content1) != string(expectedContent) {
		t.Error("First ReadFile returned incorrect content")
	}

	if string(content2) != string(expectedContent) {
		t.Error("Cached ReadFile returned incorrect content")
	}

	if string(content1) != string(content2) {
		t.Error("Cached content differs from original")
	}
}

func TestFSCache_FindFilesInDirectory(t *testing.T) {
	cache := NewFSCache(100 * time.Millisecond)
	
	tmpDir := t.TempDir()
	
	// Create some test files
	existingFiles := []string{"package.json", "yarn.lock", "README.md"}
	for _, filename := range existingFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}
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
		if result[filename] != expectedExists {
			t.Errorf("FindFilesInDirectory: expected %s to be %v, got %v", 
				filename, expectedExists, result[filename])
		}
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
		if err != nil {
			t.Fatal(err)
		}
	}

	searchFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	
	result := cache.FindFilesInDirectoryWithContent(tmpDir, searchFiles)

	// Verify file1.txt
	if !result["file1.txt"].Exists {
		t.Error("file1.txt should exist")
	}
	if string(result["file1.txt"].Content) != "content1" {
		t.Error("file1.txt content mismatch")
	}

	// Verify file2.txt
	if !result["file2.txt"].Exists {
		t.Error("file2.txt should exist")
	}
	if string(result["file2.txt"].Content) != "content2" {
		t.Error("file2.txt content mismatch")
	}

	// Verify file3.txt (non-existing)
	if result["file3.txt"].Exists {
		t.Error("file3.txt should not exist")
	}
}

func TestFSCache_TTLExpiration(t *testing.T) {
	cache := NewFSCache(50 * time.Millisecond)
	
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	
	// Write initial content
	err := os.WriteFile(tmpFile, []byte("initial"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Read file (should be cached)
	content1, err := cache.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	// Modify file
	err = os.WriteFile(tmpFile, []byte("modified"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Read again immediately (should still return cached content)
	content2, err := cache.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content1) != string(content2) {
		t.Error("Cache should return same content before TTL expiration")
	}

	// Wait for TTL expiration
	time.Sleep(100 * time.Millisecond)

	// Read again (should return new content)
	content3, err := cache.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(content3) != "modified" {
		t.Error("Cache should return new content after TTL expiration")
	}
}

func TestFSCache_ClearExpired(t *testing.T) {
	cache := NewFSCache(50 * time.Millisecond)
	
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Add some entries to cache
	cache.Stat(tmpFile)
	cache.ReadDir(tmpDir)
	cache.ReadFile(tmpFile)

	// Check initial stats
	stats1 := cache.GetStats()
	if stats1.StatCacheSize == 0 || stats1.DirCacheSize == 0 || stats1.ContentCacheSize == 0 {
		t.Error("Cache should have entries after operations")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Clear expired entries
	cache.ClearExpired()

	// Check stats after clearing
	stats2 := cache.GetStats()
	if stats2.StatCacheSize != 0 || stats2.DirCacheSize != 0 || stats2.ContentCacheSize != 0 {
		t.Error("Cache should be empty after clearing expired entries")
	}
}

func TestFSCache_Clear(t *testing.T) {
	cache := NewFSCache(1 * time.Second)
	
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Add entries to cache
	cache.Stat(tmpFile)
	cache.ReadDir(tmpDir)
	cache.ReadFile(tmpFile)

	// Verify cache has entries
	stats1 := cache.GetStats()
	if stats1.StatCacheSize == 0 || stats1.DirCacheSize == 0 || stats1.ContentCacheSize == 0 {
		t.Error("Cache should have entries")
	}

	// Clear cache
	cache.Clear()

	// Verify cache is empty
	stats2 := cache.GetStats()
	if stats2.StatCacheSize != 0 || stats2.DirCacheSize != 0 || stats2.ContentCacheSize != 0 {
		t.Error("Cache should be empty after Clear()")
	}
} 