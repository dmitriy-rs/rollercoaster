package jsmanager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
)

type PnpmWorkspace struct {
	version int
}

const pnpmLockFilename = "pnpm-lock.yaml"
const lockFileHeaderSize = 512 // Read only first 512 bytes for version detection

// Compiled regex patterns for better performance
var (
	pnpmVersionRegex = regexp.MustCompile(`lockfileVersion:\s*['"]?([^'"\\n]+)['"]?`)
)

func ParsePnpmWorkspace(dir *string) (*PnpmWorkspace, error) {
	lockFilePath := filepath.Join(*dir, pnpmLockFilename)
	if !cache.DefaultFSCache.FileExists(lockFilePath) {
		return nil, nil
	}

	// Read only the header portion for version detection
	pnpmLockFile, err := os.Open(lockFilePath)
	if err != nil {
		return nil, err
	}
	defer pnpmLockFile.Close() //nolint:errcheck

	// Read only first 512 bytes - version info is always at the top
	header := make([]byte, lockFileHeaderSize)
	n, err := pnpmLockFile.Read(header)
	if err != nil && err != io.EOF {
		return nil, err
	}

	headerStr := string(header[:n])

	// Use compiled regex to extract version
	matches := pnpmVersionRegex.FindStringSubmatch(headerStr)
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not find lockfileVersion in pnpm-lock.yaml")
	}

	version := strings.TrimSpace(matches[1])

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
