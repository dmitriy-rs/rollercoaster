package cache

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FSCache provides session-wide caching for file system operations
type FSCache struct {
	statCache    map[string]*statEntry
	dirCache     map[string]*dirEntry
	contentCache map[string]*contentEntry
	mutex        sync.RWMutex
	ttl          time.Duration
}

type statEntry struct {
	info      fs.FileInfo
	err       error
	timestamp time.Time
}

type dirEntry struct {
	entries   []fs.DirEntry
	err       error
	timestamp time.Time
}

type contentEntry struct {
	content   []byte
	err       error
	timestamp time.Time
}

// NewFSCache creates a new file system cache with the specified TTL
func NewFSCache(ttl time.Duration) *FSCache {
	return &FSCache{
		statCache:    make(map[string]*statEntry),
		dirCache:     make(map[string]*dirEntry),
		contentCache: make(map[string]*contentEntry),
		ttl:          ttl,
	}
}

// DefaultFSCache is a global cache instance with 5-second TTL
var DefaultFSCache = NewFSCache(5 * time.Second)

// isExpired checks if a cache entry has exceeded the TTL
func (c *FSCache) isExpired(timestamp time.Time) bool {
	return time.Since(timestamp) > c.ttl
}

// Stat returns file info with caching
func (c *FSCache) Stat(path string) (fs.FileInfo, error) {
	c.mutex.RLock()
	if entry, exists := c.statCache[path]; exists && !c.isExpired(entry.timestamp) {
		c.mutex.RUnlock()
		return entry.info, entry.err
	}
	c.mutex.RUnlock()

	// Cache miss or expired - fetch from file system
	info, err := os.Stat(path)

	c.mutex.Lock()
	c.statCache[path] = &statEntry{
		info:      info,
		err:       err,
		timestamp: time.Now(),
	}
	c.mutex.Unlock()

	return info, err
}

// ReadDir returns directory entries with caching
func (c *FSCache) ReadDir(dir string) ([]fs.DirEntry, error) {
	c.mutex.RLock()
	if entry, exists := c.dirCache[dir]; exists && !c.isExpired(entry.timestamp) {
		c.mutex.RUnlock()
		return entry.entries, entry.err
	}
	c.mutex.RUnlock()

	// Cache miss or expired - fetch from file system
	entries, err := os.ReadDir(dir)

	c.mutex.Lock()
	c.dirCache[dir] = &dirEntry{
		entries:   entries,
		err:       err,
		timestamp: time.Now(),
	}
	c.mutex.Unlock()

	return entries, err
}

// ReadFile returns file content with caching
func (c *FSCache) ReadFile(path string) ([]byte, error) {
	c.mutex.RLock()
	if entry, exists := c.contentCache[path]; exists && !c.isExpired(entry.timestamp) {
		c.mutex.RUnlock()
		return entry.content, entry.err
	}
	c.mutex.RUnlock()

	// Cache miss or expired - fetch from file system
	content, err := os.ReadFile(path)

	c.mutex.Lock()
	c.contentCache[path] = &contentEntry{
		content:   content,
		err:       err,
		timestamp: time.Now(),
	}
	c.mutex.Unlock()

	return content, err
}

// FileExists checks if a file exists using cached stat results
func (c *FSCache) FileExists(path string) bool {
	_, err := c.Stat(path)
	return err == nil
}

// FindFilesInDirectory batch checks for multiple files in a directory
// This is more efficient than individual FileExists calls
func (c *FSCache) FindFilesInDirectory(dir string, filenames []string) map[string]bool {
	result := make(map[string]bool)

	// Use cached directory listing if available
	entries, err := c.ReadDir(dir)
	if err != nil {
		// If directory read fails, mark all files as not found
		for _, filename := range filenames {
			result[filename] = false
		}
		return result
	}

	// Create a set of existing files for O(1) lookup
	fileSet := make(map[string]bool)
	for _, entry := range entries {
		fileSet[entry.Name()] = true
	}

	// Check which requested files exist
	for _, filename := range filenames {
		result[filename] = fileSet[filename]
	}

	return result
}

// FindFilesInDirectoryWithContent returns both existence and content for files
func (c *FSCache) FindFilesInDirectoryWithContent(dir string, filenames []string) map[string]*FileResult {
	result := make(map[string]*FileResult)

	// First, check which files exist using directory listing
	existsMap := c.FindFilesInDirectory(dir, filenames)

	// Then read content for existing files
	for filename, exists := range existsMap {
		if exists {
			fullPath := filepath.Join(dir, filename)
			content, err := c.ReadFile(fullPath)
			result[filename] = &FileResult{
				Exists:  true,
				Content: content,
				Error:   err,
			}
		} else {
			result[filename] = &FileResult{
				Exists: false,
			}
		}
	}

	return result
}

// FileResult represents the result of a file operation
type FileResult struct {
	Exists  bool
	Content []byte
	Error   error
}

// ClearExpired removes expired entries from all caches
func (c *FSCache) ClearExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	// Clear expired stat cache entries
	for path, entry := range c.statCache {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.statCache, path)
		}
	}

	// Clear expired dir cache entries
	for path, entry := range c.dirCache {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.dirCache, path)
		}
	}

	// Clear expired content cache entries
	for path, entry := range c.contentCache {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.contentCache, path)
		}
	}
}

// Clear removes all entries from the cache
func (c *FSCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.statCache = make(map[string]*statEntry)
	c.dirCache = make(map[string]*dirEntry)
	c.contentCache = make(map[string]*contentEntry)
}

// GetStats returns cache statistics
func (c *FSCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		StatCacheSize:    len(c.statCache),
		DirCacheSize:     len(c.dirCache),
		ContentCacheSize: len(c.contentCache),
		TTL:              c.ttl,
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	StatCacheSize    int
	DirCacheSize     int
	ContentCacheSize int
	TTL              time.Duration
}
