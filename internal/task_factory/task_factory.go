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
	ID          uint      `json:"id"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Tasks map[uint]Task

type TimeProvider interface {
	Now() time.Time
}

type IDGenerator interface {
	NextID() uint
}

type TaskFactory interface {
	NewTask(string, Status) (*Task, error)
}

type DefaultTimeProvider struct{}

func (rtp DefaultTimeProvider) Now() time.Time {
	return time.Now()
}

type DefaultIDGenerator struct{ Value uint }

func (idg *DefaultIDGenerator) NextID() uint {
	idg.Value++
	return idg.Value
}

type DefaultTaskFactory struct {
	Timer       TimeProvider
	IDGenerator IDGenerator
}

func (tf *DefaultTaskFactory) NewTask(description string, status Status) (*Task, error) {
	task := &Task{
		ID:          tf.IDGenerator.NextID(),
		Description: description,
		Status:      status,
		CreatedAt:   tf.Timer.Now(),
		UpdatedAt:   tf.Timer.Now(),
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
