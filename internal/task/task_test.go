package task_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	tk "github.com/alnah/task-tracker/internal/task"
	th "github.com/alnah/task-tracker/test_helpers"
)

func Test_TaskNotFoundError_Error(t *testing.T) {
	t.Run("returns a string containing the ID", func(t *testing.T) {
		err := tk.TaskNotFoundError{ID: 1}
		th.AssertErrorMessage(t, &err, err.Error(), fmt.Sprintf("%d", err.ID))
	})
}

func Test_DescriptionError_Error(t *testing.T) {
	t.Run("returns a string containing the message", func(t *testing.T) {
		err := tk.DescriptionError{"description can't be empty"}
		th.AssertErrorMessage(t, &err, err.Error(), err.Message)
	})
}

func Test_RealTimeProvider_Now(t *testing.T) {
	t.Run("returns the current time within a tolerance", func(t *testing.T) {
		timeProvider := tk.RealTimeProvider{}

		got := timeProvider.Now()
		want := time.Now()

		tolerance := 100 * time.Millisecond
		if got.Sub(want) > tolerance || want.Sub(got) > tolerance {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func Test_TaskIDGenerator_Init(t *testing.T) {
	t.Run("initializes ID generator", func(t *testing.T) {
		testTasks := th.NewTestTasks()
		testCases := []struct {
			name  string
			tasks tk.Tasks
			want  uint
		}{
			{
				name:  "happy: initializes with highest existing task ID",
				tasks: tk.Tasks{1: testTasks[1], 2: testTasks[2], 5: testTasks[3]},
				want:  5,
			},
			{
				name:  "sad: initializes with an empty tasks map",
				tasks: tk.Tasks{},
				want:  0,
			},
			{
				name:  "edge: handles empty individual tasks",
				tasks: tk.Tasks{1: tk.Task{}, 2: tk.Task{}, 3: tk.Task{}},
				want:  3,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				idGen := tk.TaskIDGenerator{}
				got := idGen.Init(tc.tasks)

				if got != tc.want {
					t.Errorf("got %v, want %v", got, tc.want)
				}
			})
		}
	})
}

func Test_TaskIDGenerator_NextID(t *testing.T) {
	t.Run("generates next ID correctly", func(t *testing.T) {
		testTasks := th.NewTestTasks()
		testCases := []struct {
			name  string
			tasks tk.Tasks
			want  uint
		}{
			{
				name:  "happy: generates next ID correctly",
				tasks: tk.Tasks{1: testTasks[1], 4: testTasks[4], 5: testTasks[5]},
				want:  6,
			},
			{
				name:  "sad: generates next ID with no tasks",
				tasks: tk.Tasks{},
				want:  1,
			},
			{
				name:  "edge: generates next ID with empty individual tasks",
				tasks: tk.Tasks{1: tk.Task{}, 2: tk.Task{}, 3: tk.Task{}},
				want:  4,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				idGen := tk.TaskIDGenerator{}
				idGen.Init(tc.tasks)

				got := idGen.NextID()

				if got != tc.want {
					t.Errorf("got %v, want %v", got, tc.want)
				}
			})
		}
	})
}

func Test_JSONFileTaskRepository_CreateTask_Happy(t *testing.T) {
	t.Run("returns a task", func(t *testing.T) {
		_, taskRepo, file := setupTaskUnitTest(t)

		wantTask := th.NewTestTask(1, "test_task_1", tk.Todo)
		gotTask, err := taskRepo.CreateTask(file.Name(), wantTask.Description)

		th.AssertNoError(t, err)
		th.AssertDeepEqual(t, gotTask, wantTask)
	})

	t.Run("calls Store.LoadData, Store.SaveData for each task created",
		func(t *testing.T) {
			mockFs, taskRepo, file := setupTaskUnitTest(t)

			numTasksToCreate := len(mockFs.Tasks)
			taskDescriptions := make([]string, numTasksToCreate)
			for i := 1; i <= numTasksToCreate; i++ {
				taskDescriptions[i-1] = fmt.Sprintf("test_task_%d", i)
			}

			for _, description := range taskDescriptions {
				_, err := taskRepo.CreateTask(file.Name(), description)
				th.AssertNoError(t, err)
			}

			wantCalls := make(Calls, 0)
			for range mockFs.Tasks {
				wantCalls = append(wantCalls, LoadData, SaveData)
			}

			th.AssertDeepEqual(t, wantCalls, mockFs.Calls)
		})
}

func Test_JSONFileTaskRepository_CreateTask_Sad_Edge(t *testing.T) {
	t.Run("returns an error", func(t *testing.T) {
		testCases := []struct {
			name      string
			loadError error
			saveError error
			desc      string
		}{
			{
				name:      "returns error context when loading fails",
				loadError: &os.PathError{},
				saveError: nil,
				desc:      "test_task",
			},
			{
				name:      "returns error context when saving fails",
				loadError: nil,
				saveError: &os.PathError{},
				desc:      "test_task",
			},
			{
				name:      "returns a DescriptionError for an empty description",
				loadError: nil,
				saveError: nil,
				desc:      "",
			},
			{
				name:      "returns a DescriptionError for a desc > 300 characters",
				loadError: nil,
				saveError: nil,
				desc:      strings.Repeat("a", 301),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockFs, taskRepo, file := setupTaskUnitTest(t)
				mockFs.LoadError = tc.loadError
				mockFs.SaveError = tc.saveError

				_, err := taskRepo.CreateTask(file.Name(), tc.desc)

				switch {
				case tc.loadError != nil || tc.saveError != nil:
					th.AssertError(t, err, &os.PathError{})
				case len(tc.desc) == 00 || len(tc.desc) > 300:
					th.AssertError(t, err, &tk.DescriptionError{})
				}
			})
		}
	})
}

func Test_JSONFileTaskRepository_UpdateTask_Happy(t *testing.T) {
	t.Run("returns an updated task", func(t *testing.T) {
		updateDescription := "update_test"
		updateStatus := tk.InProgress
		testCases := []struct {
			name   string
			id     uint
			desc   *string
			status *tk.Status
			want   tk.Task
		}{
			{
				name:   "returns a task with updated description",
				id:     1,
				desc:   &updateDescription,
				status: nil,
				want:   th.NewTestTask(1, updateDescription, tk.Todo),
			},
			{
				name:   "returns a task with updated status",
				id:     2,
				desc:   nil,
				status: &updateStatus,
				want:   th.NewTestTask(2, "test_task_2", updateStatus),
			},
			{
				name:   "returns a task with updated description and status",
				id:     3,
				desc:   &updateDescription,
				status: &updateStatus,
				want:   th.NewTestTask(3, updateDescription, updateStatus),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				_, taskRepo, file := setupTaskUnitTest(t)

				got, err := taskRepo.UpdateTask(file.Name(), tk.UpdateTaskParams{
					ID:          tc.id,
					Description: tc.desc,
					Status:      tc.status,
				})

				th.AssertNoError(t, err)
				th.AssertDeepEqual(t, got, tc.want)
			})
		}
	})

	t.Run("calls Store.LoadData and Store.SaveData for each task updated", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)

		for _, task := range mockFs.Tasks {
			updateDescription := "updated_task"
			_, err := taskRepo.UpdateTask(file.Name(), tk.UpdateTaskParams{
				ID:          task.ID,
				Description: &updateDescription,
			})
			th.AssertNoError(t, err)
		}

		wantCalls := make(Calls, 0)
		for range mockFs.Tasks {
			wantCalls = append(wantCalls, LoadData, SaveData)
		}

		th.AssertDeepEqual(t, wantCalls, mockFs.Calls)
	})
}

func Test_JSONFileTaskRepository_UpdateTask_Sad_Edge(t *testing.T) {
	t.Run("returns an error", func(t *testing.T) {
		updateDescription := "updated_task_1"
		emptyDescription := ""
		tooLongDescription := strings.Repeat("a", 301)
		testCases := []struct {
			name        string
			loadError   error
			saveError   error
			id          uint
			description *string
		}{
			{
				name:        "returns error context when loading fails",
				loadError:   &os.PathError{},
				saveError:   nil,
				id:          1,
				description: &updateDescription,
			},
			{
				name:        "returns error context when saving fails",
				loadError:   nil,
				saveError:   &os.PathError{},
				id:          1,
				description: &updateDescription,
			},
			{
				name:        "returns a DescriptionError for an empty description",
				loadError:   nil,
				saveError:   nil,
				id:          1,
				description: &emptyDescription,
			},
			{
				name:        "returns a DescriptionError for a desc > 300 characters",
				loadError:   nil,
				saveError:   nil,
				id:          1,
				description: &tooLongDescription,
			},
			{
				name:        "returns a TaskNotFoundError when ID is out of range",
				loadError:   nil,
				saveError:   nil,
				id:          0,
				description: &updateDescription,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockFs, taskRepo, file := setupTaskUnitTest(t)
				mockFs.LoadError = tc.loadError
				mockFs.SaveError = tc.saveError

				_, err := taskRepo.UpdateTask(file.Name(), tk.UpdateTaskParams{
					ID:          tc.id,
					Description: tc.description,
				})

				switch {
				case tc.loadError != nil || tc.saveError != nil:
					th.AssertError(t, err, &os.PathError{})
				case len(*tc.description) == 00 || len(*tc.description) > 300:
					th.AssertError(t, err, &tk.DescriptionError{})
				case tc.id == 0:
					th.AssertError(t, err, &tk.TaskNotFoundError{})
				}
			})
		}
	})

	t.Run("returns the original task when no updates are provided "+
		"and calls store.LoadData for each task, but doesn't call store.SaveData",
		func(t *testing.T) {
			mockFs, taskRepo, file := setupTaskUnitTest(t)
			for _, task := range mockFs.Tasks {
				_, err := taskRepo.UpdateTask(file.Name(), tk.UpdateTaskParams{
					ID: task.ID,
				})
				th.AssertNoError(t, err)
			}

			wantCalls := make(Calls, 0)
			for range mockFs.Tasks {
				wantCalls = append(wantCalls, LoadData)
			}

			th.AssertDeepEqual(t, wantCalls, mockFs.Calls)
		})
}

func Test_JSONFileTaskRepository_ReadAllTasks_Happy(t *testing.T) {
	t.Run("returns all tasks successfully", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		tasks, err := taskRepo.ReadAllTasks(file.Name())
		th.AssertNoError(t, err)
		th.AssertDeepEqual(t, tasks, mockFs.Tasks)
	})

	t.Run("calls store.LoadData once to read all tasks", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		_, err := taskRepo.ReadAllTasks(file.Name())
		th.AssertNoError(t, err)

		wantCalls := Calls{LoadData}
		th.AssertDeepEqual(t, wantCalls, mockFs.Calls)
	})
}

func Test_JSONFileTaskRepository_ReadAllTasks_Sad(t *testing.T) {
	t.Run("returns an error when loading fails", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		mockFs.LoadError = &os.PathError{}
		_, err := taskRepo.ReadAllTasks(file.Name())
		th.AssertError(t, err, mockFs.LoadError)
	})
}

func Test_JSONFileTaskRepository_ReadManyTasks_Happy(t *testing.T) {
	tasksForTest := tk.Tasks{
		1: th.NewTestTask(1, "test_task_1", tk.Todo),
		2: th.NewTestTask(2, "test_task_2", tk.Todo),
		3: th.NewTestTask(3, "test_task_3", tk.InProgress),
		4: th.NewTestTask(4, "test_task_4", tk.InProgress),
		5: th.NewTestTask(5, "test_task_5", tk.Done),
		6: th.NewTestTask(6, "test_task_6", tk.Done),
	}

	t.Run("returns filtered tasks by status", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		mockFs.Tasks = tasksForTest

		testCases := []struct {
			name   string
			status tk.Status
			want   tk.Tasks
		}{
			{
				name:   "returns todo tasks",
				status: tk.Todo,
				want:   tk.Tasks{1: mockFs.Tasks[1], 2: mockFs.Tasks[2]},
			},
			{
				name:   "returns in-progress tasks",
				status: tk.InProgress,
				want:   tk.Tasks{3: mockFs.Tasks[3], 4: mockFs.Tasks[4]},
			},
			{
				name:   "returns done tasks",
				status: tk.Done,
				want:   tk.Tasks{5: mockFs.Tasks[5], 6: mockFs.Tasks[6]},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				gotTasks, err := taskRepo.ReadManyTasks(file.Name(), tc.status)
				th.AssertNoError(t, err)
				th.AssertDeepEqual(t, gotTasks, tc.want)
			})
		}
	})

	t.Run("calls store.LoadData once", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		mockFs.Tasks = tasksForTest

		_, err := taskRepo.ReadManyTasks(file.Name(), tk.Todo)
		th.AssertNoError(t, err)

		wantCalls := Calls{LoadData}
		th.AssertDeepEqual(t, mockFs.Calls, wantCalls)
	})
}

func Test_JSONFileTaskRepository_ReadManyTasks_Sad(t *testing.T) {
	t.Run("returns an error context when loading fails", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		mockFs.LoadError = &os.PathError{}

		_, err := taskRepo.ReadManyTasks(file.Name(), tk.Todo)
		th.AssertError(t, err, mockFs.LoadError)
	})
}

func Test_JSONFileTaskRepository_ReadManyTasks_Edge(t *testing.T) {
	t.Run("returns an empty task list when no tasks match the specified status",
		func(t *testing.T) {
			_, taskRepo, file := setupTaskUnitTest(t) // all tasks are marked as todo
			gotTasks, err := taskRepo.ReadManyTasks(file.Name(), tk.Done)
			th.AssertNoError(t, err)
			th.AssertDeepEqual(t, gotTasks, tk.Tasks{})
		})
}

func Test_JSONFileTaskRepository_DeleteTask_Happy(t *testing.T) {
	t.Run("delete the specified task from the task list", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		_, err := taskRepo.DeleteTask(file.Name(), 2)
		th.AssertNoError(t, err)

		wantTasks := make(tk.Tasks)
		for id, task := range mockFs.Tasks {
			if id != 2 { // Skip the deleted task
				wantTasks[id] = task
			}
		}

		gotTasks, err := taskRepo.Store.LoadData(file.Name())
		th.AssertNoError(t, err)
		th.AssertDeepEqual(t, gotTasks, wantTasks)
	})

	t.Run("returns the deleted task", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		got, err := taskRepo.DeleteTask(file.Name(), 2)
		th.AssertNoError(t, err)
		th.AssertDeepEqual(t, got, mockFs.Tasks[2])
	})

	t.Run("calls store.LoadData and store.SaveData once", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		_, err := taskRepo.DeleteTask(file.Name(), 2)
		th.AssertNoError(t, err)

		wantCalls := Calls{LoadData, SaveData}
		th.AssertDeepEqual(t, mockFs.Calls, wantCalls)
	})

	t.Run("deletes multiple tasks and preserves task list order",
		func(t *testing.T) {
			mockFs, taskRepo, file := setupTaskUnitTest(t)

			for id := 1; id <= 8; id++ {
				if id%2 == 1 { // Delete tasks with odd IDs
					_, err := taskRepo.DeleteTask(file.Name(), uint(id))
					th.AssertNoError(t, err)
				}
			}

			wantTasks := make(tk.Tasks)
			for id, task := range mockFs.Tasks {
				if id%2 == 0 { // Only keep even tasks
					wantTasks[id] = task
				}
			}

			gotTasks, err := taskRepo.Store.LoadData(file.Name())
			th.AssertNoError(t, err)
			th.AssertDeepEqual(t, gotTasks, wantTasks)
		})
}

func Test_JSONFileTaskRepository_DeleteTask_Sad_Edge(t *testing.T) {
	t.Run("returns an error", func(t *testing.T) {
		mockFs, taskRepo, file := setupTaskUnitTest(t)
		testCases := []struct {
			name      string
			id        uint
			loadError error
			saveError error
			wantErr   error
		}{
			{
				name:      "returns a TaskNotFoundError",
				id:        0,
				loadError: nil,
				saveError: nil,
				wantErr:   &tk.TaskNotFoundError{},
			},
			{
				name:      "returns an error context when loading fails",
				id:        2,
				loadError: &os.PathError{},
				saveError: nil,
				wantErr:   &os.PathError{},
			},
			{
				name:      "returns an error context when saving fails",
				id:        2,
				loadError: nil,
				saveError: &os.PathError{},
				wantErr:   &os.PathError{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				switch {
				case tc.loadError != nil:
					mockFs.LoadError = &os.PathError{}
				case tc.saveError != nil:
					mockFs.SaveError = &os.PathError{}
				}

				_, err := taskRepo.DeleteTask(file.Name(), tc.id)
				th.AssertError(t, err, tc.wantErr)
			})
		}
	})
}

const (
	InitFile  = "InitFile"
	LoadData  = "LoadData"
	SaveData  = "SaveData"
	CloseFile = "CloseFile"
)

type Call string

type Calls []Call

type MockJSONFileStore[T tk.Tasks] struct {
	Calls     Calls
	Tasks     tk.Tasks
	LoadError error
	SaveError error
}

func (mfs *MockJSONFileStore[Tasks]) InitFile() (*os.File, error) {
	mfs.Calls = append(mfs.Calls, InitFile)
	return nil, nil
}

func (mfs *MockJSONFileStore[Tasks]) LoadData(filepath string) (Tasks, error) {
	mfs.Calls = append(mfs.Calls, LoadData)
	if mfs.LoadError != nil {
		return Tasks{}, mfs.LoadError
	}
	return Tasks(mfs.Tasks), nil
}

func (mfs *MockJSONFileStore[Tasks]) SaveData(
	tasks Tasks,
	filepath string,
) error {
	mfs.Calls = append(mfs.Calls, SaveData)
	if mfs.SaveError != nil {
		return mfs.SaveError
	}
	return nil
}

func (mfs *MockJSONFileStore[Tasks]) cleanCalls() {
	mfs.Calls = []Call{}
}

func setupTaskUnitTest(t testing.TB) (
	*MockJSONFileStore[tk.Tasks],
	*tk.JSONFileTaskRepository,
	*os.File,
) {
	t.Helper()
	mockFileStore := &MockJSONFileStore[tk.Tasks]{Tasks: th.NewTestTasks()}
	TaskRepository := &tk.JSONFileTaskRepository{
		Store:        mockFileStore,
		TimeProvider: &th.StubTimeProvider{FixedTime: th.FixedTime},
		IDGenerator:  &tk.TaskIDGenerator{},
	}

	file, err := os.CreateTemp(os.TempDir(), "test_*.json")
	th.AssertNoError(t, err)
	t.Cleanup(func() {
		os.Remove(file.Name())
		mockFileStore.cleanCalls()
	})

	return mockFileStore, TaskRepository, file
}
