package task_repository

import (
	"errors"
	"fmt"

	s "github.com/alnah/task-tracker/internal/storage"
	f "github.com/alnah/task-tracker/internal/task_factory"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	CreateTask(string) (*f.Task, error)
	UpdateTask(UpdateTaskParams) (*f.Task, error)
	DeleteTask(uint) (*f.Task, error)
	ReadManyTasks(f.Status) (f.Tasks, error)
	ReadAllTasks() (f.Tasks, error)
}

type UpdateTaskParams struct {
	ID          uint
	Description *string
	Status      *f.Status
}

type FileTaskRepository struct {
	Tasks       f.Tasks
	TaskFactory f.TaskFactory
	DataStore   s.DataStore[f.Tasks]
}

func NewFileTaskRepository(
	factory f.TaskFactory,
	dataStore s.DataStore[f.Tasks],
) (*FileTaskRepository, error) {
	repo := &FileTaskRepository{
		TaskFactory: factory,
		DataStore:   dataStore,
		Tasks:       make(f.Tasks),
	}

	if err := repo.initializeIDGenerator(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *FileTaskRepository) CreateTask(description string) (*f.Task, error) {
	if err := r.loadTasks(); err != nil {
		return &f.Task{}, fmt.Errorf("error while creating task > %w", err)
	}

	task, err := r.TaskFactory.NewTask(description, f.Todo)
	if err != nil {
		return &f.Task{}, fmt.Errorf("error while creating task > %w", err)
	}
	r.Tasks[task.ID] = *task

	if _, err = r.DataStore.SaveData(r.Tasks); err != nil {
		return &f.Task{}, fmt.Errorf("error while creating tasks > %w", err)
	}
	return task, nil
}

func (r *FileTaskRepository) UpdateTask(params UpdateTaskParams) (*f.Task, error) {
	if err := r.loadTasks(); err != nil {
		return &f.Task{}, fmt.Errorf("error while updating task > %w", err)
	}

	task, err := r.findById(params.ID)
	if err != nil {
		return &f.Task{}, fmt.Errorf("error while updating task > %w", err)
	}

	if params.Description != nil {
		task.Description = *params.Description
	}
	if params.Status != nil {
		task.Status = *params.Status
	}

	task.UpdatedAt = r.TaskFactory.Timer.Now()
	r.Tasks[params.ID] = *task
	if _, err = r.DataStore.SaveData(r.Tasks); err != nil {
		return &f.Task{}, fmt.Errorf("error while updating task > %w", err)
	}
	return task, nil
}

func (r *FileTaskRepository) DeleteTask(id uint) (*f.Task, error) {
	if err := r.loadTasks(); err != nil {
		return &f.Task{}, fmt.Errorf("error while deleting task > %w", err)
	}

	task, err := r.findById(id)
	if err != nil {
		return &f.Task{}, fmt.Errorf("error while deleting task > %w", err)
	}

	delete(r.Tasks, id)
	if _, err = r.DataStore.SaveData(r.Tasks); err != nil {
		return &f.Task{}, fmt.Errorf("error while deleting task > %w", err)
	}
	return task, nil
}

func (r *FileTaskRepository) ReadManyTasks(status f.Status) (f.Tasks, error) {
	if err := r.loadTasks(); err != nil {
		return f.Tasks{}, fmt.Errorf("error while reading tasks > %w", err)
	}

	filteredTasks := f.Tasks{}
	for _, task := range r.Tasks {
		if task.Status == status {
			filteredTasks[task.ID] = task
		}
	}
	return filteredTasks, nil
}

func (r *FileTaskRepository) ReadAllTasks() (f.Tasks, error) {
	if err := r.loadTasks(); err != nil {
		return f.Tasks{}, fmt.Errorf("error while reading all tasks > %w", err)
	}
	return r.Tasks, nil
}

func (r *FileTaskRepository) initializeIDGenerator() error {
	err := r.loadTasks()
	if err != nil {
		if errors.Is(err, s.ErrLoadingData) {
			r.Tasks = make(f.Tasks)
			r.TaskFactory.IDGenerator.Value = 0
			return nil
		}
		return fmt.Errorf("error initializing ID generator > %w", err)
	}

	var maxID uint
	for id := range r.Tasks {
		if id > maxID {
			maxID = id
		}
	}
	r.TaskFactory.IDGenerator.Value = maxID
	return nil
}

func (r *FileTaskRepository) loadTasks() error {
	tasks, err := r.DataStore.LoadData()
	if err != nil {
		return err
	}
	r.Tasks = tasks
	return nil
}

func (r *FileTaskRepository) findById(id uint) (*f.Task, error) {
	task, ok := r.Tasks[id]
	if !ok {
		return &f.Task{}, fmt.Errorf("error while finding task with ID %d > %w",
			id, ErrTaskNotFound)
	}
	return &task, nil
}
