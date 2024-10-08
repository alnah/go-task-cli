package test_helpers

import (
	tk "github.com/alnah/task-tracker/internal/task"
)

func NewTestTask(id uint, description string, status tk.Status) tk.Task {
	return tk.Task{
		ID:          id,
		Description: description,
		Status:      status,
		CreatedAt:   FixedTime,
		UpdatedAt:   FixedTime,
	}
}

func NewTestTasks() tk.Tasks {
	return tk.Tasks{
		1: NewTestTask(1, "test_task_1", tk.Todo),
		2: NewTestTask(2, "test_task_2", tk.Todo),
		3: NewTestTask(3, "test_task_3", tk.Todo),
		4: NewTestTask(4, "test_task_4", tk.Todo),
		5: NewTestTask(5, "test_task_5", tk.Todo),
		6: NewTestTask(5, "test_task_6", tk.Todo),
		7: NewTestTask(5, "test_task_7", tk.Todo),
		8: NewTestTask(5, "test_task_8", tk.Todo),
	}
}
