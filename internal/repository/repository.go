package task_repository

import (
	"errors"
	"time"
)

var EmptyDescriptionError = errors.New("description can't be empty")
var DescriptionTooLongError = errors.New("description can't exceed 300 characters")
var TaskNotFoundError = errors.New("task not found")

const (
	StatusTodo      Status = "todo"
	StatusInProcess Status = "in-process"
	StatusDone      Status = "done"
)

type Repository interface {
	ImportTasksData(tasks Tasks) Tasks
	AddTask(description string) (Task, error)
	UpdateTask(params UpdateTaskParams) (Task, error)
	DeleteTask(id uint) (bool, error)
	FindAll() Tasks
	FindMany(status Status) Tasks
}

type Timer interface {
	Now() time.Time
}

type Counter interface {
	Increment() uint
}

type RealTimer struct{}

func (t RealTimer) Now() time.Time {
	return time.Now()
}

type IDCounter struct {
	Value uint
}

func (c *IDCounter) Increment() uint {
	c.Value++
	return c.Value
}

type Status string

type Task struct {
	Description string    `json:"description"`
	ID          uint      `json:"id"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tasks map[int]Task

type UpdateTaskParams struct {
	ID          uint
	Description *string
	Status      *Status
}

type MemoTaskRepo struct {
	Timer   Timer
	Counter IDCounter
	Tasks   Tasks
}

func (r *MemoTaskRepo) ImportTasksData(tasks Tasks) Tasks {
	r.Tasks = tasks
	return r.Tasks
}

func (r *MemoTaskRepo) AddTask(description string) (Task, error) {
	if description == "" {
		return Task{}, EmptyDescriptionError
	}

	if len(description) > 300 {
		return Task{}, DescriptionTooLongError
	}

	id := r.Counter.Increment()
	now := r.Timer.Now()
	newTask := Task{
		Description: description,
		ID:          id,
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.Tasks[int(id)] = newTask
	return newTask, nil
}

func (r *MemoTaskRepo) UpdateTask(params UpdateTaskParams) (Task, error) {
	task, err := r.findById(params.ID)
	if err != nil {
		return Task{}, err
	}

	if params.Description != nil && len(*params.Description) > 300 {
		return Task{}, DescriptionTooLongError
	}

	if params.Description != nil {
		task.Description = *params.Description
	}

	if params.Status != nil {
		task.Status = Status(*params.Status)
	}

	task.UpdatedAt = r.Timer.Now()
	r.Tasks[int(params.ID)] = task
	return task, nil
}

func (r *MemoTaskRepo) DeleteTask(id uint) (bool, error) {
	_, err := r.findById(id)
	if err != nil {
		return false, err
	}

	delete(r.Tasks, int(id))
	return true, nil
}

func (r *MemoTaskRepo) FindAll() Tasks {
	return r.Tasks
}

func (r *MemoTaskRepo) FindMany(status Status) Tasks {
	result := make(Tasks)
	for id, task := range r.Tasks {
		if task.Status == status {
			result[id] = task
		}
	}
	return result
}

func (r *MemoTaskRepo) findById(id uint) (Task, error) {
	task, exists := r.Tasks[int(id)]
	if !exists {
		return Task{}, TaskNotFoundError
	}

	return task, nil
}
