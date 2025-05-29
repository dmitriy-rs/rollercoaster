package jsmanager

import (
	"fmt"
	"os/exec"
	"strings"

	configfile "github.com/dmitriy-rs/rollercoaster/internal/configFile"
	"github.com/dmitriy-rs/rollercoaster/internal/logger"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

type JsManager struct {
	manager  internalManager
	config   packageJsonConfig
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
	Scripts map[string]string `json:"scripts"`
}

const packageJsonFilename = "package.json"

func ParseJsManager(dir *string) (*JsManager, error) {
	packageJsonFile := configfile.FindInDirectory(dir, packageJsonFilename)
	if packageJsonFile == nil {
		return nil, nil
	}
	config, err := configfile.ParseFileAsJson[packageJsonConfig](packageJsonFile)
	if err != nil {
		return nil, err
	}
	manager := &JsManager{
		config:   config,
		filename: packageJsonFilename,
		manager:  nil,
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
	for name, script := range m.config.Scripts {
		tasks = append(tasks, task.Task{
			Name:        name,
			Description: script,
		})
	}
	return tasks, nil
}

func (m *JsManager) ExecuteTask(task *task.Task, args ...string) {
	cmd := m.manager.Cmd()
	manager.CommandExecute(cmd, args...)
}

func (m *JsManager) GetTitle() manager.Title {
	return manager.Title{
		Name:        m.manager.Name(),
		Description: "task runner. parsed from " + m.filename,
	}
}
