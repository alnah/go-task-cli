package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	rp "github.com/alnah/task-tracker/internal/repository"
)

var SavingTasksError = errors.New("failed to save tasks")
var LoadingTasksError = errors.New("failed to load tasks")

type Storage interface {
	SaveTasks(io.Writer, rp.Tasks) (rp.Tasks, error)
	LoadTasks(io.Reader) (rp.Tasks, error)
}

type TaskStorage struct{}

func (s TaskStorage) SaveTasks(w io.Writer, tasks rp.Tasks) (rp.Tasks, error) {
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		return rp.Tasks{}, fmt.Errorf("%w: %+v", SavingTasksError, err)
	}
	return tasks, nil
}

func (s *TaskStorage) LoadTasks(r io.Reader) (rp.Tasks, error) {
	var tasks rp.Tasks
	if err := json.NewDecoder(r).Decode(&tasks); err != nil {
		return rp.Tasks{}, fmt.Errorf("%w: %+v", LoadingTasksError, err)
	}
	return tasks, nil
}
