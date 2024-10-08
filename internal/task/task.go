package task

import (
	"fmt"
	"time"

	st "github.com/alnah/task-tracker/internal/store"
)

type Status string

const (
	Todo       Status = "todo"
	InProgress Status = "in-progress"
	Done       Status = "done"
)

type Task struct {
	ID          uint
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Tasks map[uint]Task

type UpdateTaskParams struct {
	ID          uint
	Description *string
	Status      *Status
}

type TimeProvider interface {
	Now() time.Time
}

type IDGenerator interface {
	Init(Tasks) uint
	NextID() uint
}

type TaskRepository interface {
	CreateTask(string) (Task, error)
	ReadAllTasks() (Tasks, error)
	ReadManyTasks(Status) (Tasks, error)
	UpdateTask(UpdateTaskParams) (Task, error)
	DeleteTask(uint) (Task, error)
}

type DescriptionError struct {
	Message string
}

func (e *DescriptionError) Error() string {
	return fmt.Sprintf("invalid task description: %s", e.Message)
}

type TaskNotFoundError struct {
	ID uint
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task with ID %d not found", e.ID)
}

type RealTimeProvider struct{}

func (rtp *RealTimeProvider) Now() time.Time {
	return time.Now()
}

type TaskIDGenerator struct {
	value uint
}

func (idg *TaskIDGenerator) Init(tasks Tasks) uint {
	var max uint
	for key := range tasks {
		if key > max {
			max = key
		}
	}
	idg.value = max
	return idg.value
}

func (idg *TaskIDGenerator) NextID() uint {
	idg.value++
	return idg.value
}

type JSONFileTaskRepository struct {
	Store        st.Store[Tasks]
	TimeProvider TimeProvider
	IDGenerator  IDGenerator
}

func (tr *JSONFileTaskRepository) CreateTask(
	filepath string,
	description string,
) (Task, error) {
	tasks, err := tr.Store.LoadData(filepath)
	if err != nil {
		return Task{}, fmt.Errorf("failed to load tasks data:\n>%w", err)
	}

	task, err := tr.newTask(description)
	if err != nil {
		return Task{}, fmt.Errorf("failed to build a new task:\n>%w", err)
	}

	tasks[task.ID] = task
	if err := tr.Store.SaveData(tasks, filepath); err != nil {
		return Task{}, fmt.Errorf("failed to save tasks data:\n>%w", err)
	}

	return task, nil
}

func (tr *JSONFileTaskRepository) ReadAllTasks(filepath string) (Tasks, error) {
	tasks, err := tr.Store.LoadData(filepath)
	if err != nil {
		return Tasks{}, fmt.Errorf("failed to load tasks data:\n>%w", err)
	}
	return tasks, nil
}

func (tr *JSONFileTaskRepository) ReadManyTasks(
	filepath string,
	status Status,
) (Tasks, error) {
	tasks, err := tr.ReadAllTasks(filepath)
	if err != nil {
		return Tasks{}, err
	}

	var filteredTasks = make(Tasks)
	for _, task := range tasks {
		if task.Status == status {
			filteredTasks[task.ID] = task
		}
	}

	return filteredTasks, nil
}

func (tr *JSONFileTaskRepository) UpdateTask(
	filepath string,
	update UpdateTaskParams,
) (Task, error) {
	tasks, err := tr.Store.LoadData(filepath)
	if err != nil {
		return Task{}, fmt.Errorf("failed to load tasks data:\n>%w", err)
	}

	updateTask, err := tr.findByID(tasks, update.ID)
	if err != nil {
		return Task{}, err
	}

	if update.Description == nil && update.Status == nil {
		return updateTask, nil
	}

	if update.Description != nil {
		if err := tr.validateDescription(*update.Description); err != nil {
			return Task{}, err
		}
		updateTask.Description = *update.Description
	}

	if update.Status != nil {
		updateTask.Status = *update.Status
	}

	tasks[update.ID] = updateTask
	if err := tr.Store.SaveData(tasks, filepath); err != nil {
		return Task{}, fmt.Errorf("failed to save tasks data:\n>%w", err)
	}

	return updateTask, nil
}

func (tr *JSONFileTaskRepository) DeleteTask(
	filepath string,
	id uint,
) (Task, error) {
	tasks, err := tr.Store.LoadData(filepath)
	if err != nil {
		return Task{}, fmt.Errorf("failed to load tasks data:\n>%w", err)
	}

	_, err = tr.findByID(tasks, id)
	if err != nil {
		return Task{}, err
	}

	delete(tasks, id)

	if err := tr.Store.SaveData(tasks, filepath); err != nil {
		return Task{}, fmt.Errorf("failed to save tasks data:\n>%w", err)
	}

	return tasks[id], nil
}

func (tr *JSONFileTaskRepository) newTask(desc string) (Task, error) {
	if err := tr.validateDescription(desc); err != nil {
		return Task{}, err
	}
	now := tr.TimeProvider.Now()
	return Task{
		ID:          tr.IDGenerator.NextID(),
		Description: desc,
		Status:      Todo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (tr *JSONFileTaskRepository) validateDescription(desc string) error {
	if len(desc) == 0 {
		return &DescriptionError{
			Message: fmt.Sprintf("description can't be empty"),
		}
	}
	if len(desc) > 300 {
		return &DescriptionError{
			Message: "description can't be more than 300 characters, " +
				fmt.Sprintf("but got %d characters", len(desc)),
		}
	}
	return nil
}

func (tr *JSONFileTaskRepository) findByID(tasks Tasks, id uint) (Task, error) {
	updateTask, ok := tasks[id]
	if !ok {
		return Task{}, &TaskNotFoundError{ID: id}
	}
	return updateTask, nil
}
