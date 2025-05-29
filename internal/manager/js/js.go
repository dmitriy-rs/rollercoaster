package jsmanager

import (
	configfile "github.com/dmitriy-rs/rollercoaster/internal/configFile"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

type JsManager struct {
	workspace JsWorkspace
	config    packageJsonConfig
	filename  string
}

type packageJsonConfig struct {
	Scripts map[string]string `json:"scripts"`
}

const packageJsonFilename = "package.json"

func ParseJsManager(dir *string, workspace JsWorkspace) (*JsManager, error) {
	packageJsonFile := configfile.FindInDirectory(dir, packageJsonFilename)
	if packageJsonFile == nil {
		return nil, nil
	}
	config, err := configfile.ParseFileAsJson[packageJsonConfig](packageJsonFile)
	if err != nil {
		return nil, err
	}
	manager := &JsManager{
		config:    config,
		filename:  packageJsonFilename,
		workspace: workspace,
	}

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
	cmd := m.workspace.Cmd()
	manager.CommandExecute(cmd, args...)
}

func (m *JsManager) GetTitle() manager.Title {
	return manager.Title{
		Name:        m.workspace.Name(),
		Description: "task runner. parsed from " + m.filename,
	}
}
