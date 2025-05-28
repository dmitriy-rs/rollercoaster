package manager

import (
	"fmt"
	"os"
	"path"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"

	"github.com/goccy/go-yaml"
)

type ManagerFile struct {
	filename string
	file     []byte
}

func ParseFileAsYaml[T any](mf *ManagerFile) (T, error) {
	var result T
	err := yaml.Unmarshal(mf.file, &result)
	if err != nil {
		message := fmt.Sprintf("Failed to parse YAML config %s file", mf.filename)
		logger.Error(message, err)
		os.Exit(1)
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
		os.Exit(1)
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
		filename: fullPath,
		file:     file,
	}
}
