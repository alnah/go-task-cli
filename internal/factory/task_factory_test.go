package task_factory_test

import (
	"strings"
	"testing"
	"time"

	f "github.com/alnah/task-tracker/internal/factory"
)

type MockTimer struct {
	FixedTime time.Time
}

func (t *MockTimer) Now() time.Time {
	return t.FixedTime
}

func TestTask_Validate(t *testing.T) {
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

func TestTaskFactory_NewTask(t *testing.T) {
	mockTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTimer := &MockTimer{FixedTime: mockTime}
	idGen := &f.IDGenerator{}
	factory := &f.TaskFactory{
		Timer:       mockTimer,
		IDGenerator: *idGen,
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

func TestIDGenerator_NextID(t *testing.T) {
	idGen := &f.IDGenerator{}
	for i := 1; i <= 5; i++ {
		id := idGen.NextID()
		if id != uint(i) {
			t.Errorf("expected ID %d, got %d", i, id)
		}
	}
}

func TestTaskFactory_MultipleTasks(t *testing.T) {
	mockTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	mockTimer := &MockTimer{FixedTime: mockTime}
	idGen := &f.IDGenerator{}
	factory := &f.TaskFactory{
		Timer:       mockTimer,
		IDGenerator: *idGen,
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
