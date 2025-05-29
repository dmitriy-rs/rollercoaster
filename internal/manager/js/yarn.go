package jsmanager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type YarnManager struct {
	version int
}

const yarnLockFilename = "yarn.lock"

func ParseYarnManager(dir *string) (*YarnManager, error) {
	yarnLockFile, err := os.OpenFile(path.Join(*dir, yarnLockFilename), os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil
	}
	defer yarnLockFile.Close() //nolint:errcheck

	scanner := bufio.NewScanner(yarnLockFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "@npm:") {
			return nil, fmt.Errorf("yarn 2+ lockfile unsupported")
		}
	}

	return &YarnManager{version: 1}, nil
}

func (m *YarnManager) Name() string {
	return "yarn"
}

func (m *YarnManager) Cmd() *exec.Cmd {
	return exec.Command("yarn")
}

func (m *YarnManager) InstallCmd() *exec.Cmd {
	return exec.Command("yarn", "install")
}

func (m *YarnManager) RunCmd() *exec.Cmd {
	return exec.Command("yarn", "run")
}

func (m *YarnManager) AddCmd() *exec.Cmd {
	return exec.Command("yarn", "add")
}

func (m *YarnManager) RemoveCmd() *exec.Cmd {
	return exec.Command("yarn", "remove")
}
