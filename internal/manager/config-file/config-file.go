package configfile

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"

	"github.com/goccy/go-yaml"
)

type ParseConfig struct {
	CurrentDir string
	RootDir    string
}

func (c *ParseConfig) GetDirectories() []string {
	currentDir := strings.TrimSuffix(c.CurrentDir, "/")
	rootDir := strings.TrimSuffix(c.RootDir, "/")

	if rootDir == "" {
		return []string{currentDir}
	}
	if currentDir == "" {
		// Maybe it's wrong, but we'll try to parse the root dir
		return []string{rootDir}
	}
	if currentDir == rootDir {
		return []string{rootDir}
	}

	// Pre-allocate slice based on estimated path depth
	// Count path separators to estimate capacity needed
	currentDepth := strings.Count(currentDir, string(filepath.Separator))
	rootDepth := strings.Count(rootDir, string(filepath.Separator))
	estimatedCapacity := currentDepth - rootDepth + 1
	if estimatedCapacity < 1 {
		estimatedCapacity = 2 // Minimum reasonable capacity
	}

	parsedDirectories := make([]string, 0, estimatedCapacity)
	for dir := currentDir; dir != rootDir; dir = filepath.Dir(dir) {
		parsedDirectories = append(parsedDirectories, dir)
	}
	parsedDirectories = append(parsedDirectories, rootDir)
	slices.Reverse(parsedDirectories)

	return parsedDirectories
}

type ConfigFile struct {
	Filename string
	File     []byte
}

func ParseFile[T any](mf *ConfigFile) (T, error) {
	var result T

	// Determine file type based on extension
	ext := strings.ToLower(filepath.Ext(mf.Filename))

	switch ext {
	case ".json":
		err := json.Unmarshal(mf.File, &result)
		if err != nil {
			message := fmt.Sprintf("Failed to parse JSON config %s file", mf.Filename)
			logger.Error(message, err)
			return result, err
		}
	case ".yaml", ".yml":
		err := yaml.Unmarshal(mf.File, &result)
		if err != nil {
			message := fmt.Sprintf("Failed to parse YAML config %s file", mf.Filename)
			logger.Error(message, err)
			return result, err
		}
	default:
		err := fmt.Errorf("unsupported file type: %s", ext)
		message := fmt.Sprintf("Unsupported config file type %s", mf.Filename)
		logger.Error(message, err)
		return result, err
	}

	return result, nil
}

func FindFirstInDirectory(dir *string, filenames []string) *ConfigFile {
	// Use batch file finding for better performance
	if dir == nil {
		logger.Error("Directory is nil", nil)
		return nil
	}

	resultsMap := cache.DefaultFSCache.FindFilesInDirectoryWithContent(*dir, filenames)

	// Return the first found file in the order specified
	for _, filename := range filenames {
		if result, exists := resultsMap[filename]; exists && result.Exists {
			if result.Error != nil {
				logger.Error("Failed to read file", result.Error)
				continue
			}
			fullPath := filepath.Join(*dir, filename)
			return &ConfigFile{
				Filename: fullPath,
				File:     result.Content,
			}
		}
	}
	return nil
}

func FindInDirectory(dir *string, filename string) *ConfigFile {
	if dir == nil {
		logger.Error("Directory is nil", nil)
		return nil
	}

	if filename == "" {
		logger.Error("Filename is empty", nil)
		return nil
	}

	fullPath := filepath.Join(*dir, filename)
	if !cache.DefaultFSCache.FileExists(fullPath) {
		return nil
	}

	file, err := cache.DefaultFSCache.ReadFile(fullPath)
	if err != nil {
		logger.Error("Failed to read file", err)
		return nil
	}

	return &ConfigFile{
		Filename: fullPath,
		File:     file,
	}
}
