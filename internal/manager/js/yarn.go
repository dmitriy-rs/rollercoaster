package jsmanager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
)

type YarnWorkspace struct {
	version int
}

const yarnLockFilename = "yarn.lock"

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

	scanner := bufio.NewScanner(yarnLockFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "@npm:") {
			return nil, fmt.Errorf("yarn 2+ lockfile unsupported")
		}
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
