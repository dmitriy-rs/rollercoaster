package cache

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-yaml"
)

// LRUEntry represents an entry in the LRU cache
type LRUEntry struct {
	key       string
	timestamp time.Time
	index     int // index in the heap
}

// LRUHeap implements a min-heap for LRU eviction
type LRUHeap []*LRUEntry

func (h LRUHeap) Len() int           { return len(h) }
func (h LRUHeap) Less(i, j int) bool { return h[i].timestamp.Before(h[j].timestamp) }
func (h LRUHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *LRUHeap) Push(x interface{}) {
	entry := x.(*LRUEntry)
	entry.index = len(*h)
	*h = append(*h, entry)
}

func (h *LRUHeap) Pop() interface{} {
	old := *h
	n := len(old)
	entry := old[n-1]
	entry.index = -1
	*h = old[0 : n-1]
	return entry
}

// UnifiedCache combines file system and parsed content caching with optimized concurrency
type UnifiedCache struct {
	files map[string]*CachedFile
	dirs  map[string]*CachedDir

	// Separate mutexes to reduce lock contention
	filesMutex sync.RWMutex
	dirsMutex  sync.RWMutex

	// LRU tracking for efficient eviction
	lruHeap    LRUHeap
	lruEntries map[string]*LRUEntry
	lruMutex   sync.Mutex

	ttl     time.Duration
	maxSize int
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

// NewUnifiedCache creates a new unified cache with optimized concurrency
func NewUnifiedCache(ttl time.Duration, maxSize int) *UnifiedCache {
	return &UnifiedCache{
		files:      make(map[string]*CachedFile),
		dirs:       make(map[string]*CachedDir),
		lruHeap:    make(LRUHeap, 0, maxSize),
		lruEntries: make(map[string]*LRUEntry),
		ttl:        ttl,
		maxSize:    maxSize,
	}
}

// Default cache instance
var DefaultCache = NewUnifiedCache(5*time.Second, 2000)

func (c *UnifiedCache) isExpired(timestamp time.Time) bool {
	return time.Since(timestamp) > c.ttl
}

// updateLRU updates the LRU tracking for a given key
func (c *UnifiedCache) updateLRU(key string) {
	c.lruMutex.Lock()
	defer c.lruMutex.Unlock()

	now := time.Now()
	if entry, exists := c.lruEntries[key]; exists {
		// Update existing entry
		entry.timestamp = now
		heap.Fix(&c.lruHeap, entry.index)
	} else {
		// Add new entry
		entry := &LRUEntry{
			key:       key,
			timestamp: now,
		}
		c.lruEntries[key] = entry
		heap.Push(&c.lruHeap, entry)
	}
}

// removeLRU removes a key from LRU tracking
func (c *UnifiedCache) removeLRU(key string) {
	c.lruMutex.Lock()
	defer c.lruMutex.Unlock()

	if entry, exists := c.lruEntries[key]; exists {
		heap.Remove(&c.lruHeap, entry.index)
		delete(c.lruEntries, key)
	}
}

// GetFileInfo returns file info, reading if necessary
func (c *UnifiedCache) GetFileInfo(path string) (fs.FileInfo, error) {
	c.filesMutex.RLock()
	if cached, exists := c.files[path]; exists && !c.isExpired(cached.Timestamp) {
		c.filesMutex.RUnlock()
		c.updateLRU(path)
		return cached.Info, cached.Error
	}
	c.filesMutex.RUnlock()

	// Cache miss - read from file system
	info, err := os.Stat(path)

	c.filesMutex.Lock()
	c.files[path] = &CachedFile{
		Info:      info,
		Error:     err,
		Timestamp: time.Now(),
	}
	c.filesMutex.Unlock()

	c.updateLRU(path)
	c.evictIfNeeded()

	return info, err
}

// FileExists checks if file exists using cached info
func (c *UnifiedCache) FileExists(path string) bool {
	_, err := c.GetFileInfo(path)
	return err == nil
}

// ReadFile returns file content, caching both stat and content
func (c *UnifiedCache) ReadFile(path string) ([]byte, error) {
	c.filesMutex.RLock()
	if cached, exists := c.files[path]; exists && !c.isExpired(cached.Timestamp) && cached.Content != nil {
		c.filesMutex.RUnlock()
		c.updateLRU(path)
		return cached.Content, cached.Error
	}
	c.filesMutex.RUnlock()

	// Read file and stat info together
	content, err := os.ReadFile(path)
	var info fs.FileInfo
	if err == nil {
		info, _ = os.Stat(path) // Get file info for complete cache entry
	}

	c.filesMutex.Lock()
	c.files[path] = &CachedFile{
		Info:      info,
		Content:   content,
		Error:     err,
		Timestamp: time.Now(),
	}
	c.filesMutex.Unlock()

	c.updateLRU(path)
	c.evictIfNeeded()

	return content, err
}

// ParseFile reads and parses a config file, caching the result
func (c *UnifiedCache) ParseFile(path string, target interface{}) error {
	c.filesMutex.RLock()
	if cached, exists := c.files[path]; exists && !c.isExpired(cached.Timestamp) {
		// Check if we have the parsed version cached
		if cached.Parsed != nil {
			c.filesMutex.RUnlock()
			c.updateLRU(path)
			// Copy parsed content to target using proper deep copy
			return copyParsedContent(cached.Parsed, target)
		}
		// We have content but not parsed version
		if cached.Content != nil && cached.Error == nil {
			content := cached.Content
			c.filesMutex.RUnlock()
			return c.parseAndCache(path, content, target)
		}
	}
	c.filesMutex.RUnlock()

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
	var parsed interface{}

	// Parse into a temporary value to cache
	switch ext {
	case ".json":
		var tempResult interface{}
		err = json.Unmarshal(content, &tempResult)
		if err == nil {
			parsed = tempResult
			// Also parse into the target
			err = json.Unmarshal(content, target)
		}
	case ".yaml", ".yml":
		var tempResult interface{}
		err = yaml.Unmarshal(content, &tempResult)
		if err == nil {
			parsed = tempResult
			// Also parse into the target
			err = yaml.Unmarshal(content, target)
		}
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	if err != nil {
		return err
	}

	// Update cache with parsed content
	c.filesMutex.Lock()
	if cached, exists := c.files[path]; exists {
		cached.Parsed = parsed
		cached.ParsedAs = ext
	}
	c.filesMutex.Unlock()

	return nil
}

// ReadDir returns directory entries with efficient file existence lookup
func (c *UnifiedCache) ReadDir(dir string) ([]fs.DirEntry, error) {
	c.dirsMutex.RLock()
	if cached, exists := c.dirs[dir]; exists && !c.isExpired(cached.Timestamp) {
		c.dirsMutex.RUnlock()
		c.updateLRU(dir)
		return cached.Entries, cached.Error
	}
	c.dirsMutex.RUnlock()

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

	c.dirsMutex.Lock()
	c.dirs[dir] = &CachedDir{
		Entries:   entries,
		FileMap:   fileMap,
		Error:     err,
		Timestamp: time.Now(),
	}
	c.dirsMutex.Unlock()

	c.updateLRU(dir)
	c.evictIfNeeded()

	return entries, err
}

// FindFilesInDirectory efficiently checks for multiple files
func (c *UnifiedCache) FindFilesInDirectory(dir string, filenames []string) map[string]bool {
	result := make(map[string]bool)

	c.dirsMutex.RLock()
	if cached, exists := c.dirs[dir]; exists && !c.isExpired(cached.Timestamp) && cached.FileMap != nil {
		// Use cached directory listing
		for _, filename := range filenames {
			result[filename] = cached.FileMap[filename]
		}
		c.dirsMutex.RUnlock()
		c.updateLRU(dir)
		return result
	}
	c.dirsMutex.RUnlock()

	// Cache miss - read directory
	_, err := c.ReadDir(dir)
	if err != nil {
		for _, filename := range filenames {
			result[filename] = false
		}
		return result
	}

	// Now use the cached result
	c.dirsMutex.RLock()
	if cached, exists := c.dirs[dir]; exists && cached.FileMap != nil {
		for _, filename := range filenames {
			result[filename] = cached.FileMap[filename]
		}
	}
	c.dirsMutex.RUnlock()

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

// evictIfNeeded removes old entries if cache is too large using efficient LRU
func (c *UnifiedCache) evictIfNeeded() {
	c.filesMutex.RLock()
	c.dirsMutex.RLock()
	totalEntries := len(c.files) + len(c.dirs)
	c.dirsMutex.RUnlock()
	c.filesMutex.RUnlock()

	if totalEntries <= c.maxSize {
		return
	}

	// Evict 20% of entries or at least 10
	removeCount := totalEntries / 5
	if removeCount < 10 {
		removeCount = 10
	}

	c.lruMutex.Lock()
	toRemove := make([]string, 0, removeCount)

	// Use heap to efficiently find oldest entries
	for len(toRemove) < removeCount && c.lruHeap.Len() > 0 {
		entry := heap.Pop(&c.lruHeap).(*LRUEntry)
		toRemove = append(toRemove, entry.key)
		delete(c.lruEntries, entry.key)
	}
	c.lruMutex.Unlock()

	// Remove from actual caches
	for _, key := range toRemove {
		c.filesMutex.Lock()
		delete(c.files, key)
		c.filesMutex.Unlock()

		c.dirsMutex.Lock()
		delete(c.dirs, key)
		c.dirsMutex.Unlock()
	}
}

// copyParsedContent efficiently copies cached parsed content to target using reflection
func copyParsedContent(source, target interface{}) error {
	if source == nil {
		return fmt.Errorf("source is nil")
	}

	sourceValue := reflect.ValueOf(source)
	targetValue := reflect.ValueOf(target)

	// Target must be a pointer
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetElem := targetValue.Elem()
	if !targetElem.CanSet() {
		return fmt.Errorf("target cannot be set")
	}

	// Handle different source types
	switch sourceValue.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		// For complex types, use JSON marshal/unmarshal for deep copy
		jsonData, err := json.Marshal(source)
		if err != nil {
			return fmt.Errorf("failed to marshal source: %w", err)
		}

		err = json.Unmarshal(jsonData, target)
		if err != nil {
			return fmt.Errorf("failed to unmarshal to target: %w", err)
		}
		return nil

	default:
		// For simple types, direct assignment
		if sourceValue.Type().AssignableTo(targetElem.Type()) {
			targetElem.Set(sourceValue)
			return nil
		}

		// Try conversion
		if sourceValue.Type().ConvertibleTo(targetElem.Type()) {
			targetElem.Set(sourceValue.Convert(targetElem.Type()))
			return nil
		}

		return fmt.Errorf("cannot assign %T to %T", source, target)
	}
}

// ClearExpired removes all expired entries from the cache
func (c *UnifiedCache) ClearExpired() {
	now := time.Now()
	var expiredKeys []string

	// Collect expired file entries
	c.filesMutex.RLock()
	for path, entry := range c.files {
		if now.Sub(entry.Timestamp) > c.ttl {
			expiredKeys = append(expiredKeys, path)
		}
	}
	c.filesMutex.RUnlock()

	// Collect expired directory entries
	c.dirsMutex.RLock()
	for path, entry := range c.dirs {
		if now.Sub(entry.Timestamp) > c.ttl {
			expiredKeys = append(expiredKeys, path)
		}
	}
	c.dirsMutex.RUnlock()

	// Remove expired entries
	for _, key := range expiredKeys {
		c.filesMutex.Lock()
		delete(c.files, key)
		c.filesMutex.Unlock()

		c.dirsMutex.Lock()
		delete(c.dirs, key)
		c.dirsMutex.Unlock()

		c.removeLRU(key)
	}
}

// Clear removes all entries
func (c *UnifiedCache) Clear() {
	c.filesMutex.Lock()
	c.dirsMutex.Lock()
	c.lruMutex.Lock()

	c.files = make(map[string]*CachedFile)
	c.dirs = make(map[string]*CachedDir)
	c.lruHeap = make(LRUHeap, 0, c.maxSize)
	c.lruEntries = make(map[string]*LRUEntry)

	c.lruMutex.Unlock()
	c.dirsMutex.Unlock()
	c.filesMutex.Unlock()
}

// GetStats returns cache statistics
func (c *UnifiedCache) GetStats() UnifiedCacheStats {
	c.filesMutex.RLock()
	c.dirsMutex.RLock()
	c.lruMutex.Lock()

	stats := UnifiedCacheStats{
		FileEntries: len(c.files),
		DirEntries:  len(c.dirs),
		LRUEntries:  len(c.lruEntries),
		TTL:         c.ttl,
		MaxSize:     c.maxSize,
	}

	c.lruMutex.Unlock()
	c.dirsMutex.RUnlock()
	c.filesMutex.RUnlock()

	return stats
}

// UnifiedCacheStats represents cache statistics
type UnifiedCacheStats struct {
	FileEntries int
	DirEntries  int
	LRUEntries  int
	TTL         time.Duration
	MaxSize     int
}
