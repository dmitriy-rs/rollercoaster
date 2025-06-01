package cache

import (
	"io/fs"
	"time"
)

// SimpleCache provides a simplified interface for the unified cache
// This allows existing code to migrate gradually
type SimpleCache struct {
	unified *UnifiedCache
}

// NewSimpleCache creates a wrapper around the unified cache
func NewSimpleCache() *SimpleCache {
	return &SimpleCache{
		unified: NewUnifiedCache(5*time.Second, 2000),
	}
}

// Global instance for backward compatibility
var DefaultSimpleCache = NewSimpleCache()

// DefaultFSCache is a global cache instance with 5-second TTL
// Now uses the unified cache under the hood for better efficiency
var DefaultFSCache = DefaultSimpleCache

// Deprecated: NewFSCache is kept for backward compatibility
// Use the unified cache system instead
func NewFSCache(ttl time.Duration) *SimpleCache {
	return &SimpleCache{
		unified: NewUnifiedCache(ttl, 2000),
	}
}

// File system operations
func (s *SimpleCache) Stat(path string) (fs.FileInfo, error) {
	return s.unified.GetFileInfo(path)
}

func (s *SimpleCache) ReadDir(dir string) ([]fs.DirEntry, error) {
	return s.unified.ReadDir(dir)
}

func (s *SimpleCache) ReadFile(path string) ([]byte, error) {
	return s.unified.ReadFile(path)
}

func (s *SimpleCache) FileExists(path string) bool {
	return s.unified.FileExists(path)
}

func (s *SimpleCache) FindFilesInDirectory(dir string, filenames []string) map[string]bool {
	return s.unified.FindFilesInDirectory(dir, filenames)
}

func (s *SimpleCache) FindFilesInDirectoryWithContent(dir string, filenames []string) map[string]*FileResult {
	// Convert result type to maintain compatibility
	unifiedResults := s.unified.FindFilesWithContent(dir, filenames)
	results := make(map[string]*FileResult)

	for filename, result := range unifiedResults {
		results[filename] = &FileResult{
			Exists:  result.Exists,
			Content: result.Content,
			Error:   result.Error,
		}
	}

	return results
}

// Parse operations - simplified interface
func (s *SimpleCache) ParseFile(path string, target interface{}) error {
	return s.unified.ParseFile(path, target)
}

// Maintenance operations
func (s *SimpleCache) Clear() {
	s.unified.Clear()
}

func (s *SimpleCache) ClearExpired() {
	s.unified.ClearExpired()
}

// Additional methods needed for compatibility
func (s *SimpleCache) GetFileInfo(path string) (fs.FileInfo, error) {
	return s.unified.GetFileInfo(path)
}

func (s *SimpleCache) GetStats() CacheStats {
	stats := s.unified.GetStats()
	return CacheStats{
		StatCacheSize:    stats.FileEntries,
		DirCacheSize:     stats.DirEntries,
		ContentCacheSize: stats.FileEntries, // Combined in unified cache
		TTL:              stats.TTL,
	}
}

// Backward compatibility types
type FileResult struct {
	Exists  bool
	Content []byte
	Error   error
}

type CacheStats struct {
	StatCacheSize    int
	DirCacheSize     int
	ContentCacheSize int
	TTL              time.Duration
}
