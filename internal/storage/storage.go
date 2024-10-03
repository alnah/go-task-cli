package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	r "github.com/alnah/task-tracker/internal/repository"
)

var SavingTasksError = errors.New("failed to save tasks")
var LoadingTasksError = errors.New("failed to load tasks")

type Storage interface {
	SaveTasks(io.Writer, r.Tasks) (r.Tasks, error)
	LoadTasks(io.Reader) (r.Tasks, error)
}

type StorageJSON struct {
	filepath *string
}

func (s StorageJSON) SaveTasks(writer io.Writer, tasks r.Tasks) (r.Tasks, error) {
	tasksJSON, err := json.Marshal(tasks)
	if err != nil {
		return tasks, fmt.Errorf("%w: %+v", SavingTasksError, err)
	}

	if _, err = writer.Write(tasksJSON); err != nil {
		return tasks, fmt.Errorf("%w: %+v", SavingTasksError, err)
	}

	return tasks, nil
}

func (s *StorageJSON) LoadTasks(reader io.Reader) (r.Tasks, error) {
	var tasks r.Tasks
	data, err := io.ReadAll(reader)
	if err != nil {
		return r.Tasks{}, fmt.Errorf("%w: %+v", LoadingTasksError, err)
	}

	if err = json.Unmarshal(data, &tasks); err != nil {
		return r.Tasks{}, fmt.Errorf("%w: %+v", LoadingTasksError, err)
	}

	return tasks, nil
}
