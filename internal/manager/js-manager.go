package manager

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
	"github.com/dmitriy-rs/rollercoaster/internal/ui"
)

type JsManager struct {
	manager internalManager
	config  packageJsonConfig
	filename string
}

type internalManager interface {
	Name() string
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
		config:  config,
		filename: packageJsonFilename,
		manager: nil,
	}

	managers := []internalManager{}

	pnpmManager, err := ParsePnpmManager(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	yarnManager, err := ParseYarnManager(dir)
	if err != nil {
		logger.Warning(err.Error())
	}
	npmManager, err := ParseNpmManager(dir)
	if err != nil {
		logger.Warning(err.Error())
	}

	managers = append(managers, pnpmManager, yarnManager, npmManager)

	if len(managers) == 0 {
		return nil, fmt.Errorf("no package manager found")
	}

	if len(managers) > 1 {
		formattedManagers := strings.Builder{}
		for _, manager := range managers {
			formattedManagers.WriteString(manager.Name())
			formattedManagers.WriteString(", ")
		}
		return nil, fmt.Errorf("multiple package managers found: %s", formattedManagers.String())
	}

	manager.manager = managers[0]
	return manager, nil
}

func (m *JsManager) ListTasks() ([]task.Task, error) {
	tasks := []task.Task{}
	for name, script := range m.config.scripts {
		tasks = append(tasks, task.Task{
			Name:        name,
			Description: script,
		})
	}
	return tasks, nil
}

func (m *JsManager) ExecuteTask(task *task.Task, args ...string) {
	cmd := m.manager.Cmd()
	CommandExecute(cmd, args...)
}

func (m *JsManager) GetTitle() string {
	return ui.TaskNameStyle.Render(m.manager.Name()) + ui.TextColor.Render(" task runner. parsed from "+m.filename)
}

type PnpmManager struct {
	version int
}

const pnpmLockFilename = "pnpm-lock.yaml"

func ParsePnpmManager(dir *string) (*PnpmManager, error) {
	pnpmLockFile, err := os.OpenFile(path.Join(*dir, pnpmLockFilename), os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil
	}
	defer pnpmLockFile.Close()

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

type YarnManager struct {
	version int
}

const yarnLockFilename = "yarn.lock"

func ParseYarnManager(dir *string) (*YarnManager, error) {
	yarnLockFile, err := os.OpenFile(path.Join(*dir, yarnLockFilename), os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil
	}
	defer yarnLockFile.Close()

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
