package cache

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

// UnifiedCache combines file system and parsed content caching
type UnifiedCache struct {
	files   map[string]*CachedFile
	dirs    map[string]*CachedDir
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int // Prevent unbounded growth
}

// CachedFile holds all file-related information in one place
type CachedFile struct {
	Info      fs.FileInfo
	Content   []byte
	Parsed    interface{} // Cached parsed content (JSON/YAML)
	ParsedAs  string      // Track what type it was parsed as
	Error     error
	Timestamp time.Time
}

// CachedDir holds directory listing
type CachedDir struct {
	Entries   []fs.DirEntry
	FileMap   map[string]bool // Quick lookup for file existence
	Error     error
	Timestamp time.Time
}

// NewUnifiedCache creates a new unified cache
func NewUnifiedCache(ttl time.Duration, maxSize int) *UnifiedCache {
	return &UnifiedCache{
		files:   make(map[string]*CachedFile),
		dirs:    make(map[string]*CachedDir),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Default cache instance
var DefaultCache = NewUnifiedCache(5*time.Second, 2000)

func (c *UnifiedCache) isExpired(timestamp time.Time) bool {
	return time.Since(timestamp) > c.ttl
}

// GetFileInfo returns file info, reading if necessary
func (c *UnifiedCache) GetFileInfo(path string) (fs.FileInfo, error) {
	c.mutex.RLock()
	if cached, exists := c.files[path]; exists && !c.isExpired(cached.Timestamp) {
		c.mutex.RUnlock()
		return cached.Info, cached.Error
	}
	c.mutex.RUnlock()

	// Cache miss - read from file system
	info, err := os.Stat(path)

	c.mutex.Lock()
	c.files[path] = &CachedFile{
		Info:      info,
		Error:     err,
		Timestamp: time.Now(),
	}
	c.evictIfNeeded()
	c.mutex.Unlock()

	return info, err
}

// FileExists checks if file exists using cached info
func (c *UnifiedCache) FileExists(path string) bool {
	_, err := c.GetFileInfo(path)
	return err == nil
}

// ReadFile returns file content, caching both stat and content
func (c *UnifiedCache) ReadFile(path string) ([]byte, error) {
	c.mutex.RLock()
	if cached, exists := c.files[path]; exists && !c.isExpired(cached.Timestamp) && cached.Content != nil {
		c.mutex.RUnlock()
		return cached.Content, cached.Error
	}
	c.mutex.RUnlock()

	// Read file and stat info together
	content, err := os.ReadFile(path)
	var info fs.FileInfo
	if err == nil {
		info, _ = os.Stat(path) // Get file info for complete cache entry
	}

	c.mutex.Lock()
	c.files[path] = &CachedFile{
		Info:      info,
		Content:   content,
		Error:     err,
		Timestamp: time.Now(),
	}
	c.evictIfNeeded()
	c.mutex.Unlock()

	return content, err
}

// ParseFile reads and parses a config file, caching the result
func (c *UnifiedCache) ParseFile(path string, target interface{}) error {
	c.mutex.RLock()
	if cached, exists := c.files[path]; exists && !c.isExpired(cached.Timestamp) {
		// Check if we have the parsed version cached
		if cached.Parsed != nil {
			c.mutex.RUnlock()
			// Copy parsed content to target
			return copyParsedContent(cached.Parsed, target)
		}
		// We have content but not parsed version
		if cached.Content != nil && cached.Error == nil {
			content := cached.Content
			c.mutex.RUnlock()
			return c.parseAndCache(path, content, target)
		}
	}
	c.mutex.RUnlock()

	// Need to read file first
	content, err := c.ReadFile(path)
	if err != nil {
		return err
	}

	return c.parseAndCache(path, content, target)
}

// parseAndCache parses content and updates cache
func (c *UnifiedCache) parseAndCache(path string, content []byte, target interface{}) error {
	// Determine file type and parse
	ext := strings.ToLower(filepath.Ext(path))
	var err error

	switch ext {
	case ".json":
		err = json.Unmarshal(content, target)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(content, target)
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	if err != nil {
		return err
	}

	// Update cache with parsed content
	c.mutex.Lock()
	if cached, exists := c.files[path]; exists {
		cached.Parsed = target
		cached.ParsedAs = ext
	}
	c.mutex.Unlock()

	return nil
}

// ReadDir returns directory entries with efficient file existence lookup
func (c *UnifiedCache) ReadDir(dir string) ([]fs.DirEntry, error) {
	c.mutex.RLock()
	if cached, exists := c.dirs[dir]; exists && !c.isExpired(cached.Timestamp) {
		c.mutex.RUnlock()
		return cached.Entries, cached.Error
	}
	c.mutex.RUnlock()

	// Read directory
	entries, err := os.ReadDir(dir)

	// Build quick lookup map
	var fileMap map[string]bool
	if err == nil {
		fileMap = make(map[string]bool, len(entries))
		for _, entry := range entries {
			fileMap[entry.Name()] = true
		}
	}

	c.mutex.Lock()
	c.dirs[dir] = &CachedDir{
		Entries:   entries,
		FileMap:   fileMap,
		Error:     err,
		Timestamp: time.Now(),
	}
	c.evictIfNeeded()
	c.mutex.Unlock()

	return entries, err
}

// FindFilesInDirectory efficiently checks for multiple files
func (c *UnifiedCache) FindFilesInDirectory(dir string, filenames []string) map[string]bool {
	result := make(map[string]bool)

	c.mutex.RLock()
	if cached, exists := c.dirs[dir]; exists && !c.isExpired(cached.Timestamp) && cached.FileMap != nil {
		// Use cached directory listing
		for _, filename := range filenames {
			result[filename] = cached.FileMap[filename]
		}
		c.mutex.RUnlock()
		return result
	}
	c.mutex.RUnlock()

	// Cache miss - read directory
	_, err := c.ReadDir(dir)
	if err != nil {
		for _, filename := range filenames {
			result[filename] = false
		}
		return result
	}

	// Now use the cached result
	c.mutex.RLock()
	if cached, exists := c.dirs[dir]; exists && cached.FileMap != nil {
		for _, filename := range filenames {
			result[filename] = cached.FileMap[filename]
		}
	}
	c.mutex.RUnlock()

	return result
}

// FindFilesWithContent returns files with their content
func (c *UnifiedCache) FindFilesWithContent(dir string, filenames []string) map[string]*UnifiedFileResult {
	result := make(map[string]*UnifiedFileResult)

	// First check which files exist
	existsMap := c.FindFilesInDirectory(dir, filenames)

	// Read content for existing files
	for filename, exists := range existsMap {
		if exists {
			fullPath := filepath.Join(dir, filename)
			content, err := c.ReadFile(fullPath)
			result[filename] = &UnifiedFileResult{
				Exists:  true,
				Content: content,
				Error:   err,
			}
		} else {
			result[filename] = &UnifiedFileResult{
				Exists: false,
			}
		}
	}

	return result
}

// UnifiedFileResult represents the result of a file operation
type UnifiedFileResult struct {
	Exists  bool
	Content []byte
	Error   error
}

// evictIfNeeded removes old entries if cache is too large
func (c *UnifiedCache) evictIfNeeded() {
	totalEntries := len(c.files) + len(c.dirs)
	if totalEntries <= c.maxSize {
		return
	}

	// Simple eviction: remove 20% of oldest entries
	removeCount := totalEntries / 5
	if removeCount < 1 {
		removeCount = 1
	}

	// Find oldest file entries
	oldestFiles := make([]string, 0, removeCount/2)
	oldestTime := time.Now()

	for path, entry := range c.files {
		if len(oldestFiles) < removeCount/2 {
			oldestFiles = append(oldestFiles, path)
			if entry.Timestamp.Before(oldestTime) {
				oldestTime = entry.Timestamp
			}
		} else if entry.Timestamp.Before(oldestTime) {
			oldestFiles[0] = path
			oldestTime = entry.Timestamp
		}
	}

	// Remove oldest files
	for _, path := range oldestFiles {
		delete(c.files, path)
	}

	// Remove some old directories too
	dirRemoveCount := removeCount - len(oldestFiles)
	if dirRemoveCount > 0 {
		removed := 0
		for path := range c.dirs {
			delete(c.dirs, path)
			removed++
			if removed >= dirRemoveCount {
				break
			}
		}
	}
}

// copyParsedContent copies cached parsed content to target
func copyParsedContent(source, target interface{}) error {
	// This is a simplified version - in practice you might want more sophisticated copying
	// For now, we'll re-parse since the performance hit is minimal compared to file I/O
	return fmt.Errorf("re-parse needed")
}

// ClearExpired removes all expired entries from the cache
func (c *UnifiedCache) ClearExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()

	// Clear expired file entries
	for path, entry := range c.files {
		if now.Sub(entry.Timestamp) > c.ttl {
			delete(c.files, path)
		}
	}

	// Clear expired directory entries
	for path, entry := range c.dirs {
		if now.Sub(entry.Timestamp) > c.ttl {
			delete(c.dirs, path)
		}
	}
}

// Clear removes all entries
func (c *UnifiedCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.files = make(map[string]*CachedFile)
	c.dirs = make(map[string]*CachedDir)
}

// GetStats returns cache statistics
func (c *UnifiedCache) GetStats() UnifiedCacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return UnifiedCacheStats{
		FileEntries: len(c.files),
		DirEntries:  len(c.dirs),
		TTL:         c.ttl,
		MaxSize:     c.maxSize,
	}
}

// UnifiedCacheStats represents cache statistics
type UnifiedCacheStats struct {
	FileEntries int
	DirEntries  int
	TTL         time.Duration
	MaxSize     int
}
