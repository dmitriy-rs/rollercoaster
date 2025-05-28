package manager

type Manager interface {
	ListTasks() ([]string, error)
	ExecuteTask(name string)
}

func FindManager(dir *string) (Manager, error) {
	manager, err := ParseTaskManager(dir)
	if err != nil {
		return nil, err
	}
	if manager != nil {
		return manager, nil
	}
	return nil, nil
}
