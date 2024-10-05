package task_repository

import (
	"errors"
	"fmt"

	ds "github.com/alnah/task-tracker/internal/data_store"
	tf "github.com/alnah/task-tracker/internal/task_factory"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	CreateTask(string) (*tf.Task, error)
	UpdateTask(UpdateTaskParams) (*tf.Task, error)
	DeleteTask(uint) (*tf.Task, error)
	ReadManyTasks(tf.Status) (tf.Tasks, error)
	ReadAllTasks() (tf.Tasks, error)
}

type JSONFileTaskRepository struct {
	Tasks       tf.Tasks
	TaskFactory tf.DefaultTaskFactory
	DataStore   ds.JSONFileDataStore[tf.Tasks]
}

type UpdateTaskParams struct {
	ID          uint
	Description *string
	Status      *tf.Status
}

func NewFileTaskRepository(
	factory tf.DefaultTaskFactory,
	dataStore ds.JSONFileDataStore[tf.Tasks],
) (*JSONFileTaskRepository, error) {
	repo := &JSONFileTaskRepository{
		TaskFactory: factory,
		DataStore:   dataStore,
		Tasks:       make(tf.Tasks),
	}

	if err := repo.initializeIDGenerator(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JSONFileTaskRepository) CreateTask(description string) (*tf.Task, error) {
	if err := r.loadTasks(); err != nil {
		return &tf.Task{}, fmt.Errorf("error while creating task > %w", err)
	}

	task, err := r.TaskFactory.NewTask(description, tf.Todo)
	if err != nil {
		return &tf.Task{}, fmt.Errorf("error while creating task > %w", err)
	}
	r.Tasks[task.ID] = *task

	if _, err = r.DataStore.SaveData(r.Tasks); err != nil {
		return &tf.Task{}, fmt.Errorf("error while creating tasks > %w", err)
	}
	return task, nil
}

func (r *JSONFileTaskRepository) UpdateTask(params UpdateTaskParams) (*tf.Task, error) {
	if err := r.loadTasks(); err != nil {
		return &tf.Task{}, fmt.Errorf("error while updating task > %w", err)
	}

	task, err := r.findById(params.ID)
	if err != nil {
		return &tf.Task{}, fmt.Errorf("error while updating task > %w", err)
	}

	if params.Description != nil {
		task.Description = *params.Description
	}
	if params.Status != nil {
		task.Status = *params.Status
	}

	task.UpdatedAt = r.TaskFactory.TimeProvider.Now()
	r.Tasks[params.ID] = *task
	if _, err = r.DataStore.SaveData(r.Tasks); err != nil {
		return &tf.Task{}, fmt.Errorf("error while updating task > %w", err)
	}
	return task, nil
}

func (r *JSONFileTaskRepository) DeleteTask(id uint) (*tf.Task, error) {
	if err := r.loadTasks(); err != nil {
		return &tf.Task{}, fmt.Errorf("error while deleting task > %w", err)
	}

	task, err := r.findById(id)
	if err != nil {
		return &tf.Task{}, fmt.Errorf("error while deleting task > %w", err)
	}

	delete(r.Tasks, id)
	if _, err = r.DataStore.SaveData(r.Tasks); err != nil {
		return &tf.Task{}, fmt.Errorf("error while deleting task > %w", err)
	}
	return task, nil
}

func (r *JSONFileTaskRepository) ReadManyTasks(status tf.Status) (tf.Tasks, error) {
	if err := r.loadTasks(); err != nil {
		return tf.Tasks{}, fmt.Errorf("error while reading tasks > %w", err)
	}

	filteredTasks := tf.Tasks{}
	for _, task := range r.Tasks {
		if task.Status == status {
			filteredTasks[task.ID] = task
		}
	}
	return filteredTasks, nil
}

func (r *JSONFileTaskRepository) ReadAllTasks() (tf.Tasks, error) {
	if err := r.loadTasks(); err != nil {
		return tf.Tasks{}, fmt.Errorf("error while reading all tasks > %w", err)
	}
	return r.Tasks, nil
}

func (r *JSONFileTaskRepository) initializeIDGenerator() error {
	err := r.loadTasks()
	if err != nil {
		if errors.Is(err, ds.ErrLoadingData) {
			r.Tasks = make(tf.Tasks)
			r.TaskFactory.IDGenerator.SetID(0)
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
	r.TaskFactory.IDGenerator.SetID(maxID)
	return nil
}

func (r *JSONFileTaskRepository) loadTasks() error {
	tasks, err := r.DataStore.LoadData()
	if err != nil {
		return err
	}
	r.Tasks = tasks
	return nil
}

func (r *JSONFileTaskRepository) findById(id uint) (*tf.Task, error) {
	task, ok := r.Tasks[id]
	if !ok {
		return &tf.Task{}, fmt.Errorf("error while finding task with ID %d > %w",
			id, ErrTaskNotFound)
	}
	return &task, nil
}
