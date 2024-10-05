package task_repository_test

import (
	"testing"
	"time"

	f "github.com/alnah/task-tracker/internal/task_factory"
	r "github.com/alnah/task-tracker/internal/task_repository"
)

func TestNewFileTaskRepository(t *testing.T) {
	t.Run("should make a new file task repository", func(t *testing.T) {
		factory := SpyFactory{
			Calls:       []Call{},
			Timer:       MockTimer{FixedTime: FakeTime},
			IDGenerator: MockIDGenerator{},
		}
		dataStore := SpyDataStore{[]Call{}}
		repository, err := r.NewFileTaskRepository(&factory)
	})
}

const (
	Now      = "Now"
	NextID   = "NextID"
	NewTask  = "NewTask"
	LoadData = "LoadData"
	SaveData = "SaveData"
)

var FakeTime = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

var FakeTask = f.Task{
	Description: "buy groceries",
	ID:          1,
	Status:      f.Todo,
	CreatedAt:   FakeTime,
	UpdatedAt:   FakeTime,
}

var FakeTasks = f.Tasks{1: FakeTask}

type MockTimer struct{ FixedTime time.Time }

func (mt *MockTimer) Now() time.Time {
	return mt.FixedTime
}

type MockIDGenerator struct{ Value uint }

func (mg *MockIDGenerator) NextID() uint {
	mg.Value++
	return mg.Value
}

type Call string

type SpyFactory struct {
	Calls       []Call
	Timer       *MockTimer
	IDGenerator *MockIDGenerator
}

func (sf *SpyFactory) NewTask(description string, status f.Status) (*f.Task, error) {
	return &f.Task{
		ID:          sf.IDGenerator.NextID(),
		Description: description,
		Status:      f.Todo,
		CreatedAt:   sf.Timer.Now(),
		UpdatedAt:   sf.Timer.Now(),
	}, nil
}

type SpyDataStore struct{ Calls []Call }

func (sds *SpyDataStore) SaveData(f.Task) (f.Task, error) {
	return FakeTask, nil
}

func (sds *SpyDataStore) LoadData() (f.Tasks, error) {
	return FakeTasks, nil
}
