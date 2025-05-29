package jsmanager

import (
	"os"
	"os/exec"
	"path"
)

type NpmManager struct {
}

const npmLockFilename = "package-lock.json"

func ParseNpmManager(dir *string) (*NpmManager, error) {
	packageLockFile, err := os.Stat(path.Join(*dir, npmLockFilename))
	if err != nil || packageLockFile.IsDir() {
		return nil, nil
	}
	return &NpmManager{}, nil
}

func (m *NpmManager) Name() string {
	return "npm"
}

func (m *NpmManager) Cmd() *exec.Cmd {
	return exec.Command("npm")
}

func (m *NpmManager) InstallCmd() *exec.Cmd {
	return exec.Command("npm", "install")
}

func (m *NpmManager) RunCmd() *exec.Cmd {
	return exec.Command("npm", "run")
}

func (m *NpmManager) AddCmd() *exec.Cmd {
	return exec.Command("npm", "i")
}

func (m *NpmManager) RemoveCmd() *exec.Cmd {
	return exec.Command("npm", "uninstall")
}
