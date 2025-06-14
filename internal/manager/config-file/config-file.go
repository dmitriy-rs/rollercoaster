package configfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"

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

	parsedDirectories := []string{}
	for dir := currentDir; dir != rootDir; dir = filepath.Dir(dir) {
		parsedDirectories = append(parsedDirectories, dir)
	}
	parsedDirectories = append(parsedDirectories, rootDir)
	slices.Reverse(parsedDirectories)

	fmt.Println(parsedDirectories)

	return parsedDirectories
}

type ConfigFile struct {
	Filename string
	File     []byte
}

func ParseFileAsYaml[T any](mf *ConfigFile) (T, error) {
	var result T
	err := yaml.Unmarshal(mf.File, &result)
	if err != nil {
		message := fmt.Sprintf("Failed to parse YAML config %s file", mf.Filename)
		logger.Error(message, err)
		return result, err
	}
	return result, nil
}

func ParseFileAsJson[T any](mf *ConfigFile) (T, error) {
	var result T
	err := json.Unmarshal(mf.File, &result)
	if err != nil {
		message := fmt.Sprintf("Failed to parse JSON config %s file", mf.Filename)
		logger.Error(message, err)
		return result, err
	}
	return result, nil
}

func FindFirstInDirectory(dir *string, filenames []string) *ConfigFile {
	for _, filename := range filenames {
		file := FindInDirectory(dir, filename)
		if file != nil {
			return file
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
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.ReadFile(fullPath)
	if err != nil {
		logger.Error("Failed to read file", err)
		return nil
	}

	return &ConfigFile{
		Filename: fullPath,
		File:     file,
	}
}
