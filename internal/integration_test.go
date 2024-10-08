package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	st "github.com/alnah/task-tracker/internal/store"
	tk "github.com/alnah/task-tracker/internal/task"
	th "github.com/alnah/task-tracker/test_helpers"
)

// TODO: add a get filepath method for store?
// TODO: tester le store

// Edge
// abrupt shutdowns
// recovery procedures

func Test_Integration_Happy(t *testing.T) {
	t.Parallel()

	t.Run("creates multiple tasks successfully", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)

		var wantTasks = tk.Tasks{}
		var gotTasks = tk.Tasks{}

		for i := 1; i <= repetitions; i++ {
			id := uint(i)
			desc := getTaskDesc(id)

			gotTask, err := taskRepo.CreateTask(filepath, desc)
			th.AssertNoError(t, err)

			gotTasks[id] = gotTask
			wantTasks[id] = th.NewTestTask(id, desc, tk.Todo)
		}

		th.AssertDeepEqual(t, gotTasks, wantTasks)
	})

	t.Run("updates multiple tasks successfully", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)

		var wantTasks = tk.Tasks{}
		var gotTasks = tk.Tasks{}

		for i := 1; i <= repetitions; i++ {
			id := uint(i)
			desc := getTaskDesc(id)
			updateDesc := fmt.Sprintf("update_%s", desc)
			var updateStatus tk.Status = tk.Done // status update defaults to "done"
			if i%2 == 0 {
				updateStatus = tk.InProgress // even tasks are updated to "in-progress"
			}

			_, err := taskRepo.CreateTask(filepath, desc)
			th.AssertNoError(t, err)

			updatedTask, err := taskRepo.UpdateTask(filepath, tk.UpdateTaskParams{
				ID:          id,
				Description: &updateDesc,
				Status:      &updateStatus,
			})
			th.AssertNoError(t, err)

			gotTasks[id] = updatedTask
			wantTasks[id] = th.NewTestTask(id, updateDesc, updateStatus)
		}

		th.AssertDeepEqual(t, gotTasks, wantTasks)
	})

	t.Run("deletes multiple tasks successfully", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)

		var wantTasks = tk.Tasks{}
		var gotTasks = tk.Tasks{}

		for i := 1; i <= repetitions; i++ {
			id := uint(i)
			desc := getTaskDesc(id)

			gotTask, err := taskRepo.CreateTask(filepath, desc)
			th.AssertNoError(t, err)

			if i%2 == 0 { // delete even tasks
				gotTasks[id] = gotTask
				wantTasks[id] = th.NewTestTask(id, desc, tk.Todo)
			}
		}

		th.AssertDeepEqual(t, gotTasks, wantTasks)
	})

	t.Run("reads all tasks successfully", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)

		var wantTasks = tk.Tasks{}

		for i := 1; i <= repetitions; i++ {
			id := uint(i)
			desc := getTaskDesc(id)

			wantTask, err := taskRepo.CreateTask(filepath, desc)
			th.AssertNoError(t, err)
			wantTasks[id] = wantTask
		}

		gotTasks, err := taskRepo.ReadAllTasks(filepath)
		th.AssertNoError(t, err)
		th.AssertDeepEqual(t, gotTasks, wantTasks)
	})

	t.Run("reads multiple tasks by status successfully", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)

		gotTasksByStatus := map[tk.Status]tk.Tasks{}
		wantTasksByStatus := map[tk.Status]tk.Tasks{}

		for _, status := range []tk.Status{tk.Todo, tk.InProgress, tk.Done} {
			gotTasksByStatus[status] = tk.Tasks{}
			wantTasksByStatus[status] = tk.Tasks{}
		}

		for i := 1; i <= repetitions; i++ {
			id := uint(i)
			desc := getTaskDesc(id)

			var updatedStatus tk.Status
			switch i % 3 {
			case 0:
				updatedStatus = tk.Todo
			case 1:
				updatedStatus = tk.InProgress
			case 2:
				updatedStatus = tk.Done
			}

			_, err := taskRepo.CreateTask(filepath, desc)
			th.AssertNoError(t, err)

			gotTask, err := taskRepo.UpdateTask(filepath, tk.UpdateTaskParams{
				ID:          id,
				Description: &desc,
				Status:      &updatedStatus,
			})
			th.AssertNoError(t, err)

			wantTasksByStatus[updatedStatus][id] = gotTask
		}

		for status := range gotTasksByStatus {
			tasks, err := taskRepo.ReadManyTasks(filepath, status)
			th.AssertNoError(t, err)
			gotTasksByStatus[status] = tasks
		}

		th.AssertDeepEqual(t, gotTasksByStatus, wantTasksByStatus)
	})
}

func Test_Integration_Sad(t *testing.T) {
	testCases := []struct {
		name        string
		storeParams storeParams
		wantErr     error
	}{
		{
			name: "returns an os.PathError when directory creation fails " +
				"due to a long path",
			storeParams: storeParams{
				DestDir:  strings.Repeat("a", 1000),
				Filename: "test_tasks.json",
				InitData: st.EmptyObject,
			},
			wantErr: &os.PathError{},
		},
		{
			name: "returns an os.PathError when file creation fails " +
				"due to a long filename",
			storeParams: storeParams{
				DestDir:  t.TempDir(),
				Filename: strings.Repeat("a", 1000) + ".json",
				InitData: st.EmptyObject,
			},
			wantErr: &os.PathError{},
		},
		{
			name: "returns a FilenameExtError when filename lacks " +
				"the .json extension",
			storeParams: storeParams{
				DestDir:  t.TempDir(),
				Filename: "bad_filename.incorrect",
				InitData: st.EmptyObject,
			},
			wantErr: &st.FilenameExtError{},
		},
		{
			name: "returns an InitDataError when initial data is malformed",
			storeParams: storeParams{
				DestDir:  t.TempDir(),
				Filename: "test_tasks.json",
				InitData: "|",
			},
			wantErr: &st.InitDataError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := setupJSONFileStore(t, tc.storeParams)
			_, err := store.InitFile()
			th.AssertError(t, err, tc.wantErr)
		})
	}

	t.Run("fails to load data from the store", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)
		_, err := taskRepo.CreateTask(filepath, "test_task")
		th.AssertNoError(t, err)

		_, err = taskRepo.Store.LoadData("bad_file.json")
		th.AssertError(t, err, &os.PathError{})
	})

	t.Run("fails to save data to the store", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)
		_, err := taskRepo.CreateTask(filepath, "test_task")
		th.AssertNoError(t, err)

		tasks, err := taskRepo.ReadAllTasks(filepath)
		th.AssertNoError(t, err)

		err = taskRepo.Store.SaveData(tasks, "bad_file.json")
		th.AssertError(t, err, &os.PathError{})
	})

	t.Run("validates task descriptions", func(t *testing.T) {
		testCases := []struct {
			name        string
			description string
			wantErr     error
		}{
			{
				name: "returns a DescriptionError when adding a task with " +
					"an empty description",
				description: "",
				wantErr:     &tk.DescriptionError{},
			},
			{
				name: "returns a DescriptionError when adding a task with " +
					"a too long description",
				description: strings.Repeat("a", 301),
				wantErr:     &tk.DescriptionError{},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				taskRepo, filepath := setupTaskRepository(t)

				_, err := taskRepo.CreateTask(filepath, tc.description)
				th.AssertError(t, err, tc.wantErr)
			})
		}
	})

	t.Run("handles update task error conditions", func(t *testing.T) {
		taskRepo, filepath := setupTaskRepository(t)
		updateDescription := "update_test_task"

		emptyString := new(string)
		longDescription := strings.Repeat("a", 301)

		_, err := taskRepo.CreateTask(filepath, "test_task")
		th.AssertNoError(t, err)

		tests := []struct {
			name        string
			id          uint
			description *string
			wantErr     interface{}
		}{
			{
				name: "returns a DescriptionError when updating a task with " +
					"an empty description",
				id:          1,
				description: emptyString,
				wantErr:     &tk.DescriptionError{},
			},
			{
				name: "returns a DescriptionError when updating a task with " +
					"a too long description",
				id:          1,
				description: &longDescription,
				wantErr:     &tk.DescriptionError{},
			},
			{
				name: "returns a TaskNotFoundError when updating a task with " +
					"a non-existing ID",
				id:          0,
				description: &updateDescription,
				wantErr:     &tk.TaskNotFoundError{},
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				updateParams := tk.UpdateTaskParams{
					ID:          tc.id,
					Description: tc.description,
				}

				var expectedError error
				if tc.wantErr != nil {
					expectedError = tc.wantErr.(error)
				}
				_, err := taskRepo.UpdateTask(filepath, updateParams)
				th.AssertError(t, err, expectedError)
			})
		}
	})

	t.Run("returns a TaskNotFoundError when deleting a task with a non-existing ID",
		func(t *testing.T) {
			taskRepo, filepath := setupTaskRepository(t)

			_, err := taskRepo.DeleteTask(filepath, 0)
			th.AssertError(t, err, &tk.TaskNotFoundError{})
		})
}

const repetitions = 100

type storeParams struct {
	DestDir  string
	Filename string
	InitData st.JSONInitData
}

func setupJSONFileStore(
	t testing.TB,
	params storeParams,
) st.JSONFileStore[tk.Tasks] {
	t.Helper()
	store := st.JSONFileStore[tk.Tasks]{
		DestDir:  params.DestDir,
		Filename: params.Filename,
		InitData: params.InitData,
	}
	t.Cleanup(func() { os.RemoveAll(store.DestDir) })
	return store
}

func setupTaskRepository(t testing.TB) (tk.JSONFileTaskRepository, string) {
	t.Helper()
	tempDir := t.TempDir()
	// setup store
	store := st.JSONFileStore[tk.Tasks]{
		DestDir:  tempDir,
		Filename: "integration_test_tasks.json",
		InitData: st.EmptyObject,
	}
	filepath := filepath.Join(store.DestDir, store.Filename)
	t.Cleanup(func() { os.RemoveAll(store.DestDir) })

	_, err := store.InitFile()
	th.AssertNoError(t, err)

	// setup stub time provider
	timeProvider := th.StubTimeProvider{FixedTime: th.FixedTime}

	// setup task id generator
	idGenerator := tk.TaskIDGenerator{}
	tasks, err := store.LoadData(filepath)
	th.AssertNoError(t, err)
	idGenerator.Init(tasks)

	// setup JSON file task repository
	taskRepository := tk.JSONFileTaskRepository{
		Store:        &store,
		TimeProvider: &timeProvider,
		IDGenerator:  &idGenerator,
	}

	return taskRepository, filepath
}

func getTaskDesc(id uint) string {
	return fmt.Sprintf("test_task_%d", id)
}
