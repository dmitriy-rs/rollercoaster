package jsmanager

import (
	"os/exec"
	"path/filepath"

	"github.com/dmitriy-rs/rollercoaster/internal/manager/cache"
)

type BunWorkspace struct {
}

const bunLockFilename = "bun.lockb"
const bunLockTextFilename = "bun.lock"

func ParseBunWorkspace(dir *string) (*BunWorkspace, error) {
	// Check for both binary and text lock file formats
	binaryLockFile := filepath.Join(*dir, bunLockFilename)
	textLockFile := filepath.Join(*dir, bunLockTextFilename)

	if !cache.DefaultFSCache.FileExists(binaryLockFile) && !cache.DefaultFSCache.FileExists(textLockFile) {
		return nil, nil
	}
	return &BunWorkspace{}, nil
}

func GetDefaultBunWorkspace() BunWorkspace {
	return BunWorkspace{}
}

func (m *BunWorkspace) Name() string {
	return "bun"
}

func (m *BunWorkspace) ExecName() string {
	return "bunx"
}

func (m *BunWorkspace) Cmd() *exec.Cmd {
	return exec.Command("bun", "run")
}

func (m *BunWorkspace) InstallCmd() *exec.Cmd {
	return exec.Command("bun", "install")
}

func (m *BunWorkspace) ExecuteCmd() *exec.Cmd {
	return exec.Command("bunx")
}

func (m *BunWorkspace) AddCmd() *exec.Cmd {
	return exec.Command("bun", "add")
}

func (m *BunWorkspace) RemoveCmd() *exec.Cmd {
	return exec.Command("bun", "remove")
}
