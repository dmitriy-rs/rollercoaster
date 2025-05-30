package jsmanager

import (
	"os"
	"os/exec"
	"path/filepath"
)

type NpmWorkspace struct {
}

const npmLockFilename = "package-lock.json"

func ParseNpmWorkspace(dir *string) (*NpmWorkspace, error) {
	packageLockFile, err := os.Stat(filepath.Join(*dir, npmLockFilename))
	if err != nil || packageLockFile.IsDir() {
		return nil, nil
	}
	return &NpmWorkspace{}, nil
}

func (m *NpmWorkspace) Name() string {
	return "npm"
}

func (m *NpmWorkspace) ExecName() string {
	return "npx"
}

func (m *NpmWorkspace) Cmd() *exec.Cmd {
	return exec.Command("npm", "run")
}

func (m *NpmWorkspace) InstallCmd() *exec.Cmd {
	return exec.Command("npm", "install")
}

func (m *NpmWorkspace) ExecuteCmd() *exec.Cmd {
	return exec.Command("npx")
}

func (m *NpmWorkspace) AddCmd() *exec.Cmd {
	return exec.Command("npm", "i")
}

func (m *NpmWorkspace) RemoveCmd() *exec.Cmd {
	return exec.Command("npm", "uninstall")
}
