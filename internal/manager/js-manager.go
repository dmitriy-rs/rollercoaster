package manager

import "os/exec"

type JsManager struct {
	manager internalManager
	config  packageJsonConfig
}

type internalManager interface {
	Cmd() *exec.Cmd
	RunCmd() *exec.Cmd
	InstallCmd() *exec.Cmd
	AddCmd() *exec.Cmd
	RemoveCmd() *exec.Cmd
}

type packageJsonConfig struct {
	scripts map[string]string `json:"scripts"`
}

const packageJsonFilename = "package.json"

func ParseJsManager(dir *string) (*JsManager, error) {
	packageJsonFile := FindInDirectory(dir, packageJsonFilename)
	if packageJsonFile == nil {
		return nil, nil
	}
	config, err := ParseFileAsJson[packageJsonConfig](packageJsonFile)
	if err != nil {
		return nil, err
	}
	manager := &JsManager{
		config: config,
		manager: nil,
	}

	return manager, nil
}

type PnpmManager struct {
	version int
}

const pnpmLockFilename = "pnpm-lock.yaml"

func ParsePnpmManager(dir *string) (*PnpmManager, error) {
	return nil, nil
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

type YarnManager struct {
	version int
}

const yarnLockFilename = "yarn.lock"

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


type NpmManager struct {
}

const npmLockFilename = "package-lock.json"

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