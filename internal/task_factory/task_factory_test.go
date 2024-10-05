package task_factory_test

import (
	"strings"
	"testing"
	"time"

	f "github.com/alnah/task-tracker/internal/task_factory"
)

func TestRealTimeProvider_Now(t *testing.T) {
	t.Run("should return the time now", func(t *testing.T) {
		timer := f.DefaultTimeProvider{}
		if time.Now().Truncate(time.Second) != timer.Now().Truncate(time.Second) {
			t.Errorf("expected %v, got %v", time.Now().Truncate(time.Second),
				timer.Now().Truncate(time.Second))
		}
	})
}

func TestDefaultTaskFactory_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		task    f.Task
		wantErr error
	}{
		{
			name: "valid task",
			task: f.Task{
				Description: "Test task",
				Status:      f.Todo,
			},
			wantErr: nil,
		},
		{
			name: "empty description",
			task: f.Task{
				Description: "",
				Status:      f.Todo,
			},
			wantErr: f.ErrEmptyDescription,
		},
		{
			name: "too long description",
			task: f.Task{
				Description: strings.Repeat("a", 301),
				Status:      f.Todo,
			},
			wantErr: f.ErrTooLongDescription,
		},
		{
			name: "invalid status",
			task: f.Task{
				Description: "Test task",
				Status:      "invalid",
			},
			wantErr: f.ErrBadStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.task.Validate()
			if err != tc.wantErr {
				t.Errorf("expected error %v, got %v", err, tc.wantErr)
			}
		})
	}
}

func TestDefaultTaskFactory_NewTask(t *testing.T) {
	factory := &f.DefaultTaskFactory{
		TimeProvider: mockTimer,
		IDGenerator:  &f.DefaultIDGenerator{},
	}

	testCases := []struct {
		name        string
		description string
		status      f.Status
		wantErr     error
		wantID      uint
	}{
		{
			name:        "valid task",
			description: "Test task",
			status:      f.Todo,
			wantErr:     nil,
			wantID:      1,
		},
		{
			name:        "empty description",
			description: "",
			status:      f.Todo,
			wantErr:     f.ErrEmptyDescription,
			wantID:      0,
		},
		{
			name:        "too long description",
			description: strings.Repeat("a", 301),
			status:      f.Todo,
			wantErr:     f.ErrTooLongDescription,
			wantID:      0,
		},
		{
			name:        "invalid status",
			description: "Test task",
			status:      "invalid",
			wantErr:     f.ErrBadStatus,
			wantID:      0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			task, err := factory.NewTask(tc.description, tc.status)
			if err != tc.wantErr {
				t.Errorf("expected error %v, got %v", tc.wantErr, err)
			}

			if err == nil {
				if task.ID != tc.wantID {
					t.Errorf("expected ID %d, got %d", tc.wantID, task.ID)
				}
				if !task.CreatedAt.Equal(mockTime) {
					t.Errorf("expected CreatedAt %v, got %v", mockTime, task.CreatedAt)
				}
				if !task.UpdatedAt.Equal(mockTime) {
					t.Errorf("expected UpdatedAt %v, got %v", mockTime, task.UpdatedAt)
				}
			} else {
				if task != nil {
					t.Errorf("expected task to be nil when error occurs")
				}
			}
		})
	}
}

func TestDefaultIDGenerator_SetID(t *testing.T) {
	idGen := &f.DefaultIDGenerator{}
	testCases := []struct {
		input    uint
		expected uint
	}{
		{0, 0},
		{100, 100},
		{1000, 1000},
	}

	for _, tc := range testCases {
		if idGen.SetID(tc.input) != tc.expected {
			t.Errorf("expected ID to be %d, got %d", tc.expected, idGen.Value)
		}
	}
}

func TestDefaultIDGenerator_NextID(t *testing.T) {
	idGen := &f.DefaultIDGenerator{}
	for i := 1; i <= 5; i++ {
		id := idGen.NextID()
		if id != uint(i) {
			t.Errorf("expected ID %d, got %d", i, id)
		}
	}
}

func TestDefaultTaskFactory_MultipleTasks(t *testing.T) {
	idGen := &f.DefaultIDGenerator{}
	factory := &f.DefaultTaskFactory{
		TimeProvider: mockTimer,
		IDGenerator:  idGen,
	}

	descriptions := []string{"Task 1", "Task 2", "Task 3"}
	for i, desc := range descriptions {
		task, err := factory.NewTask(desc, f.Todo)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expectedID := uint(i + 1)
		if task.ID != expectedID {
			t.Errorf("expected ID %d, got %d", expectedID, task.ID)
		}
		if task.Description != desc {
			t.Errorf("expected Description %q, got %q", desc, task.Description)
		}
		if !task.CreatedAt.Equal(mockTime) {
			t.Errorf("expected CreatedAt %v, got %v", mockTime, task.CreatedAt)
		}
		if !task.UpdatedAt.Equal(mockTime) {
			t.Errorf("expected UpdatedAt %v, got %v", mockTime, task.UpdatedAt)
		}
	}
}

var mockTime = time.Date(2006, time.January, 2, 15, 4, 5, 0, time.UTC)
var mockTimer = &MockTimeProvider{FixedTime: mockTime}

type MockTimeProvider struct {
	FixedTime time.Time
}

func (t *MockTimeProvider) Now() time.Time {
	return t.FixedTime
}
