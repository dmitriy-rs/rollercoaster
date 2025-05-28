package manager

import (
	"errors"
	"os/exec"
	"strings"
)

type TaskManager struct {
	version string
	tasks   []string
}

type TaskManagerConfig struct {
	Version string              `yaml:"version"`
	Tasks   map[string]struct{} `yaml:"tasks"`
}

var localTaskFilenames = [4]string{
	"Taskfile.yml",
	"taskfile.yml",
	"Taskfile.yaml",
	"taskfile.yaml",
}
var distTaskFilenames = [4]string{
	"Taskfile.dist.yml",
	"taskfile.dist.yml",
	"Taskfile.dist.yaml",
	"taskfile.dist.yaml",
}

func ParseTaskManager(dir *string) (*TaskManager, error) {
	tasks := []string{}

	localFile := FindFirstInDirectory(dir, localTaskFilenames[:])
	distFile := FindFirstInDirectory(dir, distTaskFilenames[:])

	if localFile != nil {
		if err := populateTasksFromFile(localFile, &tasks); err != nil {
			return nil, err
		}
	}
	if distFile != nil {
		if err := populateTasksFromFile(distFile, &tasks); err != nil {
			return nil, err
		}
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	return &TaskManager{
		version: "3.0",
		tasks:   tasks,
	}, nil
}

func populateTasksFromFile(file *ManagerFile, tasks *[]string) error {
	config, err := parseConfig(file)
	if err != nil {
		return err
	}

	for task := range config.Tasks {
		*tasks = append(*tasks, task)
	}

	return nil
}

func parseConfig(file *ManagerFile) (*TaskManagerConfig, error) {
	config, err := ParseFileAsYaml[TaskManagerConfig](file)
	if err != nil {
		return nil, err
	}

	if config.Version == "" {
		return nil, errors.New("Taskfile version is not specified: " + file.filename)
	}

	if !strings.HasPrefix(config.Version, "3") {
		return nil, errors.New("Unsupported Taskfile version: " + config.Version)
	}

	return &config, nil
}

func (tm *TaskManager) ListTasks() ([]string, error) {
	if tm.tasks == nil {
		return nil, nil
	}
	return tm.tasks, nil
}

func (tm *TaskManager) ExecuteTask(name string) {
	cmd := exec.Command("task", name)
	TaskExecute(cmd)
}
