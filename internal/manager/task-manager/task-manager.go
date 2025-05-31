package taskmanager

import (
	"errors"
	"os/exec"
	"strings"

	config "github.com/dmitriy-rs/rollercoaster/internal/manager/config-file"
	"github.com/dmitriy-rs/rollercoaster/internal/manager"
	"github.com/dmitriy-rs/rollercoaster/internal/task"
)

type TaskManager struct {
	config    *TaskManagerConfig
	filenames []string
}

type taskMap = map[string]struct {
	Description string `yaml:"desc"`
}

type TaskManagerConfig struct {
	Version string  `yaml:"version"`
	Tasks   taskMap `yaml:"tasks"`
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
	tm := &TaskManager{}

	localFile := config.FindFirstInDirectory(dir, localTaskFilenames[:])
	distFile := config.FindFirstInDirectory(dir, distTaskFilenames[:])

	if distFile != nil {
		config, err := parseConfig(distFile)
		if err != nil {
			return nil, err
		}
		tm.config = config
		tm.filenames = append(tm.filenames, distFile.Filename)
	}
	if localFile != nil {
		config, err := parseConfig(localFile)

		if err != nil {
			return nil, err
		}
		if tm.config == nil {
			tm.config = config
		} else {
			for taskName, task := range config.Tasks {
				tm.config.Tasks[taskName] = task
			}
		}
		tm.filenames = append(tm.filenames, localFile.Filename)
	}
	if tm.config == nil {
		return nil, nil
	}
	return tm, nil
}

func parseConfig(file *config.ConfigFile) (*TaskManagerConfig, error) {
	config, err := config.ParseFileAsYaml[TaskManagerConfig](file)
	if err != nil {
		return nil, err
	}

	if config.Version == "" {
		return nil, errors.New("Taskfile version is not specified: " + file.Filename)
	}

	if !strings.HasPrefix(config.Version, "3") {
		return nil, errors.New("Unsupported Taskfile version: " + config.Version)
	}

	return &config, nil
}

func (tm *TaskManager) ListTasks() ([]task.Task, error) {
	if tm.config.Tasks == nil {
		return nil, nil
	}
	tasks := []task.Task{}
	for name, taskInfo := range tm.config.Tasks {
		tasks = append(tasks, task.Task{
			Name:        name,
			Description: taskInfo.Description,
		})
	}
	task.SortTasks(tasks)
	return tasks, nil
}

func (tm *TaskManager) ExecuteTask(task *task.Task, args ...string) {
	cmd := exec.Command("task", task.Name)
	manager.CommandExecute(cmd, args...)
}

func (tm *TaskManager) GetTitle() manager.Title {
	if len(tm.filenames) == 0 {
		return manager.Title{
			Name:        "task",
			Description: "task runner",
		}
	}
	return manager.Title{
		Name:        "task",
		Description: "parsed from " + strings.Join(tm.filenames, ", "),
	}
}
