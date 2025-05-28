package manager

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

	localFile := FindFirstInDirectory(dir, localTaskFilenames[:])
	distFile := FindFirstInDirectory(dir, distTaskFilenames[:])

	if distFile != nil {
		config, err := parseConfig(distFile)
		if err != nil {
			return nil, err
		}
		tm.config = config
		tm.filenames = append(tm.filenames, distFile.filename)
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
		tm.filenames = append(tm.filenames, localFile.filename)
	}
	if tm.config == nil {
		return nil, nil
	}
	return tm, nil
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

func (tm *TaskManager) ExecuteTask(task task.Task) {
	cmd := exec.Command("task", task.Name)
	TaskExecute(cmd)
}

var (
	taskNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#43aba2")).
			Bold(true)
	textColor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3f3f3f"))
)

func (tm *TaskManager) GetTitle() string {
	if len(tm.filenames) == 0 {
		return taskNameStyle.Render("task") + textColor.Render(" task runner")
	}
	return taskNameStyle.Render("task") + textColor.Render(" task runner. parsed from "+strings.Join(tm.filenames, ", "))
}
