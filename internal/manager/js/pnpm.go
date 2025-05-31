package jsmanager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PnpmWorkspace struct {
	version int
}

const pnpmLockFilename = "pnpm-lock.yaml"

func ParsePnpmWorkspace(dir *string) (*PnpmWorkspace, error) {
	pnpmLockFile, err := os.OpenFile(filepath.Join(*dir, pnpmLockFilename), os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil
	}
	defer pnpmLockFile.Close() //nolint:errcheck

	firstLine, err := bufio.NewReader(pnpmLockFile).ReadString('\n')
	if err != nil {
		return nil, nil
	}

	firstLine = strings.TrimSpace(firstLine)
	version := strings.TrimPrefix(firstLine, "lockfileVersion: '")
	version = strings.TrimSuffix(version, "'")

	switch version {
	case "9.0":
		return &PnpmWorkspace{version: 10}, nil
	case "6.0":
		return &PnpmWorkspace{version: 9}, nil
	default:
		return nil, fmt.Errorf("unsupported pnpm lockfile version: %s", version)
	}

}

func GetDefaultPnpmWorkspace() PnpmWorkspace {
	// check if pnpm is installed and it's version to parse
	return PnpmWorkspace{version: 10}
}

func (m *PnpmWorkspace) Name() string {
	switch m.version {
	case 10:
		return "pnpm@10+"
	case 9:
		return "pnpm@9+"
	default:
		return "pnpm"
	}
}

func (m *PnpmWorkspace) ExecName() string {
	return "pnpx"
}

func (m *PnpmWorkspace) Cmd() *exec.Cmd {
	return exec.Command("pnpm", "run")
}

func (m *PnpmWorkspace) ExecuteCmd() *exec.Cmd {
	return exec.Command("pnpx")
}

func (m *PnpmWorkspace) InstallCmd() *exec.Cmd {
	return exec.Command("pnpm", "install")
}

func (m *PnpmWorkspace) AddCmd() *exec.Cmd {
	return exec.Command("pnpm", "add")
}

func (m *PnpmWorkspace) RemoveCmd() *exec.Cmd {
	return exec.Command("pnpm", "remove")
}
