package task_factory

import (
	"errors"
	"time"
)

var ErrEmptyDescription = errors.New("description cannot be empty")
var ErrTooLongDescription = errors.New("description can't exceed 300 characters")
var ErrBadStatus = errors.New("status must be one of: todo, in-progress, done")

type Status string

const (
	Todo       Status = "todo"
	InProgress Status = "in-progress"
	Done       Status = "done"
)

type Task struct {
	Description string    `json:"description"`
	ID          uint      `json:"id"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tasks map[uint]Task

type Timer interface {
	Now() time.Time
}

type RealTimer struct{}

func (t RealTimer) Now() time.Time {
	return time.Now()
}

type IDGenerator struct{ Value uint }

func (g *IDGenerator) NextID() uint {
	g.Value++
	return g.Value
}

type TaskFactory struct {
	Timer       Timer
	IDGenerator IDGenerator
}

func (f *TaskFactory) NewTask(description string, status Status) (*Task, error) {
	task := &Task{
		ID:          f.IDGenerator.NextID(),
		Description: description,
		Status:      status,
		CreatedAt:   f.Timer.Now(),
		UpdatedAt:   f.Timer.Now(),
	}

	if err := task.Validate(); err != nil {
		return nil, err
	}

	return task, nil
}

func (t *Task) Validate() error {
	if len(t.Description) == 0 {
		return ErrEmptyDescription
	}

	if len(t.Description) > 300 {
		return ErrTooLongDescription
	}

	if t.Status != Todo && t.Status != InProgress && t.Status != Done {
		return ErrBadStatus
	}

	return nil
}
