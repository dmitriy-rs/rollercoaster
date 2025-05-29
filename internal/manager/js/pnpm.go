package jsmanager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type PnpmManager struct {
	version int
}

const pnpmLockFilename = "pnpm-lock.yaml"

func ParsePnpmManager(dir *string) (*PnpmManager, error) {
	pnpmLockFile, err := os.OpenFile(path.Join(*dir, pnpmLockFilename), os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil
	}
	defer pnpmLockFile.Close() //nolint:errcheck

	firstLine, err := bufio.NewReader(pnpmLockFile).ReadString('\n')
	if err != nil {
		return nil, nil
	}

	version := strings.TrimPrefix(firstLine, "lockfileVersion: '")
	version = strings.TrimSuffix(version, "'")
	version = strings.TrimSpace(version)

	switch version {
	case "9.0":
		return &PnpmManager{version: 10}, nil
	case "6.0":
		return &PnpmManager{version: 9}, nil
	default:
		return nil, fmt.Errorf("unsupported pnpm lockfile version: %s", version)
	}

}

func (m *PnpmManager) Name() string {
	return "pnpm"
}

func (m *PnpmManager) Cmd() *exec.Cmd {
	return exec.Command("pnpm")
}

func (m *PnpmManager) RunCmd() *exec.Cmd {
	return exec.Command("pnpm", "run")
}

func (m *PnpmManager) InstallCmd() *exec.Cmd {
	return exec.Command("pnpm", "install")
}

func (m *PnpmManager) AddCmd() *exec.Cmd {
	return exec.Command("pnpm", "add")
}

func (m *PnpmManager) RemoveCmd() *exec.Cmd {
	return exec.Command("pnpm", "remove")
}
