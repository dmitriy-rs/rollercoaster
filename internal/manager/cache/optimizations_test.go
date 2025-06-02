package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentAccess tests the separated mutex optimization
func TestConcurrentAccess(t *testing.T) {
	cache := NewUnifiedCache(1*time.Second, 100)
	tmpDir := t.TempDir()

	// Create test files
	for i := 0; i < 5; i++ {
		content := fmt.Sprintf(`{"test": %d}`, i)
		err := os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.json", i)), []byte(content), 0644)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Start many concurrent operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				// Mix file and directory operations
				fileNum := (id + j) % 5
				filePath := filepath.Join(tmpDir, fmt.Sprintf("file%d.json", fileNum))

				// File operations
				_, err := cache.ReadFile(filePath)
				if err != nil {
					errors <- fmt.Errorf("ReadFile error: %w", err)
					return
				}

				// Directory operations
				_, err = cache.ReadDir(tmpDir)
				if err != nil {
					errors <- fmt.Errorf("ReadDir error: %w", err)
					return
				}

				// Parse operations
				var result map[string]interface{}
				err = cache.ParseFile(filePath, &result)
				if err != nil {
					errors <- fmt.Errorf("ParseFile error: %w", err)
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent access failed: %v", err)
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}

	stats := cache.GetStats()
	assert.Greater(t, stats.FileEntries, 0, "Should have cached file entries")
	assert.Greater(t, stats.DirEntries, 0, "Should have cached directory entries")
	assert.Greater(t, stats.LRUEntries, 0, "Should have LRU entries")
}

// TestLRUEviction tests the efficient LRU eviction algorithm
func TestLRUEviction(t *testing.T) {
	// Small cache to trigger eviction
	cache := NewUnifiedCache(1*time.Hour, 5)
	tmpDir := t.TempDir()

	// Create more files than cache capacity
	testFiles := make([]string, 10)
	for i := 0; i < 10; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		testFiles[i] = filepath.Join(tmpDir, filename)
		err := os.WriteFile(testFiles[i], []byte(fmt.Sprintf("content %d", i)), 0644)
		require.NoError(t, err)
	}

	// Read files in order to populate cache
	for i := 0; i < 7; i++ { // Read 7 files to exceed capacity
		_, err := cache.ReadFile(testFiles[i])
		require.NoError(t, err)
	}

	stats := cache.GetStats()
	assert.LessOrEqual(t, stats.FileEntries, 5, "Cache should not exceed max size")
	assert.LessOrEqual(t, stats.LRUEntries, 5, "LRU should not exceed max size")

	// The first files should have been evicted
	cache.filesMutex.RLock()
	_, exists := cache.files[testFiles[0]]
	cache.filesMutex.RUnlock()
	assert.False(t, exists, "First file should have been evicted")

	// Recent files should still be cached
	cache.filesMutex.RLock()
	_, exists = cache.files[testFiles[6]]
	cache.filesMutex.RUnlock()
	assert.True(t, exists, "Recent file should still be cached")
}

// TestParseCacheEfficiency tests that parsed content is properly cached and copied
func TestParseCacheEfficiency(t *testing.T) {
	cache := NewUnifiedCache(1*time.Hour, 100)
	tmpDir := t.TempDir()

	// Create a JSON file
	testData := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"config": map[string]interface{}{
			"enabled": true,
			"count":   42,
		},
	}

	jsonData, err := json.Marshal(testData)
	require.NoError(t, err)

	jsonFile := filepath.Join(tmpDir, "config.json")
	err = os.WriteFile(jsonFile, jsonData, 0644)
	require.NoError(t, err)

	// First parse should read from file and cache the result
	var result1 map[string]interface{}
	err = cache.ParseFile(jsonFile, &result1)
	require.NoError(t, err)
	assert.Equal(t, "test", result1["name"])
	assert.Equal(t, "1.0.0", result1["version"])

	// Verify content is cached
	cache.filesMutex.RLock()
	cached, exists := cache.files[jsonFile]
	cache.filesMutex.RUnlock()
	require.True(t, exists, "File should be cached")
	assert.NotNil(t, cached.Parsed, "Parsed content should be cached")
	assert.Equal(t, ".json", cached.ParsedAs, "Parse type should be recorded")

	// Second parse should use cached parsed content
	var result2 map[string]interface{}
	err = cache.ParseFile(jsonFile, &result2)
	require.NoError(t, err)
	assert.Equal(t, result1, result2, "Cached parse should return same result")

	// Modify result1 to ensure deep copy works
	result1["name"] = "modified"

	// Third parse should still return original data
	var result3 map[string]interface{}
	err = cache.ParseFile(jsonFile, &result3)
	require.NoError(t, err)
	assert.Equal(t, "test", result3["name"], "Cached data should not be affected by modifications")
}

// TestYAMLParseCaching tests YAML parsing and caching
func TestYAMLParseCaching(t *testing.T) {
	cache := NewUnifiedCache(1*time.Hour, 100)
	tmpDir := t.TempDir()

	// Create a YAML file
	yamlContent := `
name: test-yaml
version: 2.0.0
config:
  enabled: true
  count: 100
  items:
    - first
    - second
    - third
`

	yamlFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(yamlFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	// Parse YAML
	var result map[string]interface{}
	err = cache.ParseFile(yamlFile, &result)
	require.NoError(t, err)

	assert.Equal(t, "test-yaml", result["name"])
	assert.Equal(t, "2.0.0", result["version"])

	// Verify cached
	cache.filesMutex.RLock()
	cached, exists := cache.files[yamlFile]
	cache.filesMutex.RUnlock()
	require.True(t, exists, "YAML file should be cached")
	assert.NotNil(t, cached.Parsed, "Parsed YAML should be cached")
	assert.Contains(t, []string{".yaml", ".yml"}, cached.ParsedAs, "Parse type should be YAML")

	// Second parse should use cache
	var result2 map[string]interface{}
	err = cache.ParseFile(yamlFile, &result2)
	require.NoError(t, err)
	assert.Equal(t, result["name"], result2["name"], "Cached YAML parse should work")
}

// BenchmarkLRUOperations tests LRU performance
func BenchmarkLRUOperations(b *testing.B) {
	cache := NewUnifiedCache(1*time.Hour, 1000)

	b.Run("UpdateLRU", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i%500) // Reuse some keys
			cache.updateLRU(key)
		}
	})

	b.Run("EvictionWithLRU", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1200; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.updateLRU(key)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cache.evictIfNeeded()
		}
	})
}
