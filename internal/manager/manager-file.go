package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"

	"github.com/goccy/go-yaml"
)

type ManagerFile struct {
	Filename string
	File     []byte
}

func ParseFileAsYaml[T any](mf *ManagerFile) (T, error) {
	var result T
	err := yaml.Unmarshal(mf.File, &result)
	if err != nil {
		message := fmt.Sprintf("Failed to parse YAML config %s file", mf.Filename)
		logger.Error(message, err)
		return result, err
	}
	return result, nil
}

func ParseFileAsJson[T any](mf *ManagerFile) (T, error) {
	var result T
	err := json.Unmarshal(mf.File, &result)
	if err != nil {
		message := fmt.Sprintf("Failed to parse JSON config %s file", mf.Filename)
		logger.Error(message, err)
		return result, err
	}
	return result, nil
}

func FindFirstInDirectory(dir *string, filenames []string) *ManagerFile {
	for _, filename := range filenames {
		file := FindInDirectory(dir, filename)
		if file != nil {
			return file
		}
	}
	return nil
}

func FindInDirectory(dir *string, filename string) *ManagerFile {
	if dir == nil {
		logger.Error("Directory is nil", nil)
		return nil
	}

	if filename == "" {
		logger.Error("Filename is empty", nil)
		return nil
	}

	fullPath := path.Join(*dir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.ReadFile(fullPath)
	if err != nil {
		logger.Error("Failed to read file", err)
		return nil
	}

	return &ManagerFile{
		Filename: fullPath,
		File:     file,
	}
}
