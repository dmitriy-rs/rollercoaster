package jsmanager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
)

type YarnWorkspace struct {
	version int
}

const yarnLockFilename = "yarn.lock"
const yarnHeaderSize = 1024 // Read first 1KB for yarn version detection

// Compiled regex patterns for better performance
var (
	yarnV2Regex = regexp.MustCompile(`@npm:`)
)

func ParseYarnWorkspace(dir *string) (*YarnWorkspace, error) {
	filename := filepath.Join(*dir, yarnLockFilename)
	if !cache.DefaultFSCache.FileExists(filename) {
		return nil, nil
	}

	yarnLockFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer yarnLockFile.Close() //nolint:errcheck

	// Read only first 1KB for version detection - @npm: pattern appears early
	header := make([]byte, yarnHeaderSize)
	n, err := yarnLockFile.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}

	headerStr := string(header[:n])

	// Use compiled regex to check for yarn 2+ pattern
	if yarnV2Regex.MatchString(headerStr) {
		return nil, fmt.Errorf("yarn 2+ lockfile unsupported")
	}

	return &YarnWorkspace{version: 1}, nil
}

func GetDefaultYarnWorkspace() YarnWorkspace {
	return YarnWorkspace{version: 1}
}

func (m *YarnWorkspace) Name() string {
	return "yarn"
}

func (m *YarnWorkspace) ExecName() string {
	return "yarn run"
}

func (m *YarnWorkspace) Cmd() *exec.Cmd {
	return exec.Command("yarn")
}

func (m *YarnWorkspace) InstallCmd() *exec.Cmd {
	return exec.Command("yarn", "install")
}

func (m *YarnWorkspace) ExecuteCmd() *exec.Cmd {
	return exec.Command("yarn", "run")
}

func (m *YarnWorkspace) AddCmd() *exec.Cmd {
	return exec.Command("yarn", "add")
}

func (m *YarnWorkspace) RemoveCmd() *exec.Cmd {
	return exec.Command("yarn", "remove")
}
