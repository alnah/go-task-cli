package repository_test

import (
	"errors"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	r "github.com/alnah/task-tracker/internal/repository"
)

func TestCounterIncrement(t *testing.T) {
	t.Run("should increment from the value 1", func(t *testing.T) {
		counter := r.IDCounter{}
		assertUInt(t, counter.Increment(), 1)
	})

	t.Run("should increment multiple times", func(t *testing.T) {
		counter := r.IDCounter{}
		for i := 1; i <= 10; i++ {
			assertUInt(t, counter.Increment(), uint(i))
		}
	})
}

func TestTimerNow(t *testing.T) {
	t.Run("should return the current time", func(t *testing.T) {
		assertTime(t, r.RealTimer{}.Now(), time.Now())
	})
}

func TestImportTasksData(t *testing.T) {
	testCases := []struct {
		name      string
		tasksData r.Tasks
	}{
		{
			name:      "should import tasks data",
			tasksData: getTestTasks(),
		},
		{
			name:      "should handle empty tasks data",
			tasksData: r.Tasks{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			repository := newTestRepository(nil)
			importedTasks := repository.ImportTasksData(tt.tasksData)
			assertTasks(t, importedTasks, tt.tasksData)
		})
	}
}

func TestAddTask(t *testing.T) {
	t.Run("should append a task with correct data to the tasks list",
		func(t *testing.T) {
			repository := newTestRepository(nil)
			got, err := repository.AddTask("buy groceries")

			assertNoError(t, err)
			want := buildTask("buy groceries", 1, r.StatusTodo)
			assertTask(t, got, want)
		})

	t.Run("should increment task IDs sequentially", func(t *testing.T) {
		repository := newTestRepository(nil)
		descriptions := []string{"task 1", "task 2", "task 3"}

		for i, desc := range descriptions {
			task, err := repository.AddTask(desc)
			assertNoError(t, err)
			assertUInt(t, task.ID, uint(i+1))
		}

		assertInt(t, len(repository.Tasks), len(descriptions))
	})

	t.Run("should handle duplicate descriptions without errors",
		func(t *testing.T) {
			repository := newTestRepository(nil)
			descriptions := []string{"task 1", "task 1", "task 1"}

			for _, desc := range descriptions {
				_, err := repository.AddTask(desc)
				assertNoError(t, err)
			}

			assertInt(t, len(repository.Tasks), len(descriptions))
		})

	t.Run("should handle empty description", func(t *testing.T) {
		repository := newTestRepository(nil)
		_, err := repository.AddTask("")
		assertError(t, err, r.EmptyDescriptionError)
	})

	t.Run("should restrict task description to 300 characters",
		func(t *testing.T) {
			repository := newTestRepository(nil)
			longDescription := strings.Repeat("a", 301)
			_, err := repository.AddTask(longDescription)
			assertError(t, err, r.DescriptionTooLongError)
		})
}
func TestUpdateTask(t *testing.T) {
	todo := func() *r.Status { s := r.StatusTodo; return &s }()
	inProcess := func() *r.Status { s := r.StatusInProcess; return &s }()
	done := func() *r.Status { s := r.StatusDone; return &s }()

	newDescription := func() *string { d := "cook dinner"; return &d }()

	t.Run("should correctly update task properties based on various scenarios",
		func(t *testing.T) {
			updateTestCases := []struct {
				name        string
				id          uint
				description *string
				status      *r.Status
			}{
				{
					name:   "should update task status only",
					id:     1,
					status: todo,
				},
				{
					name:        "should update task description only",
					id:          1,
					description: newDescription,
				},
				{
					name:        "should update both task status and description",
					id:          1,
					status:      inProcess,
					description: newDescription,
				},
				{
					name:   `should update task status to "todo"`,
					id:     1,
					status: todo,
				},
				{
					name:   `should update task status to "in-process"`,
					id:     1,
					status: inProcess,
				},
				{
					name:   `should update task status to "done"`,
					id:     1,
					status: done,
				},
				{
					name: "should handle task update with empty description and status",
					id:   1,
				},
			}

			repository := r.TaskRepository{stubTimer, r.IDCounter{}, r.Tasks{1: t1}}
			for _, tt := range updateTestCases {
				t.Run(tt.name, func(t *testing.T) {
					got, err := repository.UpdateTask(r.UpdateTaskParams{
						ID:          tt.id,
						Description: tt.description,
						Status:      tt.status,
					})

					assertNoError(t, err)
					want := r.Task{
						ID:          tt.id,
						Description: repository.Tasks[1].Description,
						Status:      repository.Tasks[1].Status,
						CreatedAt:   repository.Tasks[1].CreatedAt,
						UpdatedAt:   stubTimer.Now(),
					}

					if tt.description != nil {
						want.Description = *tt.description
					}

					if tt.status != nil {
						want.Status = *tt.status
					}
					assertTask(t, got, want)
				})
			}
		})

	t.Run("should handle task update for non-existent task", func(t *testing.T) {
		tasks := r.Tasks{1: buildTask("buy groceries", 1, r.StatusTodo)}
		repository := newTestRepository(tasks)
		_, err := repository.UpdateTask(r.UpdateTaskParams{
			ID:     2,
			Status: inProcess,
		})
		assertError(t, err, r.TaskNotFoundError)
	})

	t.Run("should handle duplicate descriptions without errors",
		func(t *testing.T) {
			repository := newTestRepository(r.Tasks{
				1: buildTask("buy groceries", 1, r.StatusTodo),
				2: buildTask("cook dinner", 2, r.StatusTodo),
			})
			newDescription := "cook dinner"
			got, err := repository.UpdateTask(r.UpdateTaskParams{
				ID:          1,
				Description: &newDescription,
			})
			assertNoError(t, err)
			want := buildTask("cook dinner", 1, r.StatusTodo)
			want.CreatedAt = repository.Tasks[1].CreatedAt
			want.UpdatedAt = stubTimer.Now()
			assertTask(t, got, want)
		})

	t.Run("should restrict task description to 300 characters", func(t *testing.T) {
		tasks := r.Tasks{1: buildTask("buy groceries", 1, r.StatusTodo)}
		repository := newTestRepository(tasks)
		longDescription := strings.Repeat("a", 301)
		_, err := repository.UpdateTask(r.UpdateTaskParams{
			ID:          1,
			Description: &longDescription,
		})
		assertError(t, err, r.DescriptionTooLongError)
	})
}

func TestDeleteTask(t *testing.T) {
	tests := []struct {
		name          string
		initialTasks  r.Tasks
		taskToDelete  uint
		expectedTasks r.Tasks
		expectedError error
	}{
		{
			name:          "should delete the first task",
			initialTasks:  getTestTasks(),
			taskToDelete:  1,
			expectedTasks: removeTask(getTestTasks(), 1),
		},
		{
			name:          "should delete the second task",
			initialTasks:  getTestTasks(),
			taskToDelete:  2,
			expectedTasks: removeTask(getTestTasks(), 2),
		},
		{
			name:          "should delete the third task",
			initialTasks:  getTestTasks(),
			taskToDelete:  3,
			expectedTasks: removeTask(getTestTasks(), 3),
		},
		{
			name:          "should handle task delete for non-existent task",
			initialTasks:  r.Tasks{1: buildTask("buy groceries", 1, r.StatusTodo)},
			taskToDelete:  2,
			expectedTasks: r.Tasks{1: buildTask("buy groceries", 1, r.StatusTodo)},
			expectedError: r.TaskNotFoundError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := newTestRepository(tt.initialTasks)
			_, err := repository.DeleteTask(tt.taskToDelete)

			if tt.expectedError != nil {
				assertError(t, err, tt.expectedError)
			} else {
				assertNoError(t, err)
			}

			assertTasks(t, repository.Tasks, tt.expectedTasks)
		})
	}

	t.Run("should delete all tasks in random order", func(t *testing.T) {
		tasks := getTestTasks()
		repository := newTestRepository(tasks)
		tasksToDelete := []uint{1, 2, 3, 4, 5}
		rand.Shuffle(len(tasksToDelete), func(i, j int) {
			tasksToDelete[i], tasksToDelete[j] = tasksToDelete[j], tasksToDelete[i]
		})

		for _, id := range tasksToDelete {
			_, err := repository.DeleteTask(id)
			assertNoError(t, err)
		}

		assertInt(t, len(repository.Tasks), 0)
	})
}

func TestFindAll(t *testing.T) {
	t.Run("should find all tasks", func(t *testing.T) {
		tasks := getTestTasks()
		repository := newTestRepository(tasks)
		assertTasks(t, repository.FindAll(), tasks)
	})

	t.Run("should find no task when the repository is empty", func(t *testing.T) {
		repository := newTestRepository(nil)
		assertTasks(t, repository.FindAll(), r.Tasks{})
	})
}

func TestFindMany(t *testing.T) {
	tasks := getTestTasks()
	repository := newTestRepository(tasks)
	tests := []struct {
		name     string
		status   r.Status
		expected r.Tasks
	}{
		{
			name:     `should find all "todo" tasks`,
			status:   r.StatusTodo,
			expected: filterTasksByStatus(tasks, r.StatusTodo),
		},
		{
			name:     `should find all "in-process" tasks`,
			status:   r.StatusInProcess,
			expected: filterTasksByStatus(tasks, r.StatusInProcess),
		},
		{
			name:     `should find all "done" tasks`,
			status:   r.StatusDone,
			expected: filterTasksByStatus(tasks, r.StatusDone),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertTasks(t, repository.FindMany(tt.status), tt.expected)
		})
	}

	t.Run("should find nothing when no tasks of a certain status are found",
		func(t *testing.T) {
			tasks := filterTasksByStatus(getTestTasks(), r.StatusInProcess)
			repository := newTestRepository(tasks)
			assertTasks(t, repository.FindMany(r.StatusDone), r.Tasks{})
		})

	t.Run("should find no task when the repository is empty", func(t *testing.T) {
		repository := newTestRepository(nil)
		assertTasks(t, repository.FindMany(r.StatusTodo), r.Tasks{})
	})
}

var stubTimer = StubTimer{
	fixedTime: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
}

type StubTimer struct {
	fixedTime time.Time
}

func (t StubTimer) Now() time.Time {
	return t.fixedTime
}

var (
	t1 = buildTask("buy groceries", 1, r.StatusTodo)
	t2 = buildTask("cook dinner", 2, r.StatusTodo)
	t3 = buildTask("do laundry", 3, r.StatusInProcess)
	t4 = buildTask("write report", 4, r.StatusInProcess)
	t5 = buildTask("attend meeting", 5, r.StatusDone)
)

func newTestRepository(tasks r.Tasks) *r.TaskRepository {
	if tasks == nil {
		tasks = r.Tasks{}
	}
	return &r.TaskRepository{
		Timer:   stubTimer,
		Counter: r.IDCounter{},
		Tasks:   tasks,
	}
}

func getTestTasks() r.Tasks {
	return r.Tasks{
		1: buildTask("buy groceries", 1, r.StatusTodo),
		2: buildTask("cook dinner", 2, r.StatusTodo),
		3: buildTask("do laundry", 3, r.StatusInProcess),
		4: buildTask("write report", 4, r.StatusInProcess),
		5: buildTask("attend meeting", 5, r.StatusDone),
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal("got an error but didn't want one")
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func assertUInt(t testing.TB, got, want uint) {
	t.Helper()
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func assertInt(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func assertTime(t testing.TB, got, want time.Time) {
	t.Helper()
	const tolerance = 100 * time.Millisecond // tolerance for comparison
	if got.Sub(want) > tolerance || want.Sub(got) > tolerance {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func assertTask(t testing.TB, got, want r.Task) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func assertTasks(t testing.TB, got, want r.Tasks) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func buildTask(description string, id uint, status r.Status) r.Task {
	return r.Task{
		Description: description,
		ID:          id,
		Status:      status,
		CreatedAt:   stubTimer.Now(),
		UpdatedAt:   stubTimer.Now(),
	}
}

func removeTask(tasks r.Tasks, id uint) r.Tasks {
	newTasks := make(r.Tasks)
	for k, v := range tasks {
		if k != int(id) {
			newTasks[k] = v
		}
	}
	return newTasks
}

func filterTasksByStatus(tasks r.Tasks, status r.Status) r.Tasks {
	filtered := make(r.Tasks)
	for id, task := range tasks {
		if task.Status == status {
			filtered[id] = task
		}
	}
	return filtered
}
