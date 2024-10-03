package service_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	r "github.com/alnah/task-tracker/internal/repository"
	s "github.com/alnah/task-tracker/internal/service"
)

func TestAddTask(t *testing.T) {
	t.Run("should add a task while verifying storage and repository interactions", func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		task, err := service.AddTask("buy groceries")

		assertNoError(t, err)
		assertTask(t, task, r.Task{})
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks, storageSaveTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryAddTask,
			repositoryFindAll,
		})
	})

	t.Run("should return error when storage.LoadTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.loadTasksErr = io.ErrClosedPipe

		_, err := service.AddTask("buy groceries")
		assertError(t, err, io.ErrClosedPipe)
	})

	t.Run("should return error when repository.AddTask fails", func(t *testing.T) {
		service, _, spyRepository := newTestService()
		spyRepository.addTaskErr = io.ErrClosedPipe

		_, err := service.AddTask("buy groceries")
		assertError(t, err, io.ErrClosedPipe)
	})

	t.Run("should return error when storage.SaveTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.saveTasksErr = io.ErrClosedPipe

		_, err := service.AddTask("buy groceries")
		assertError(t, err, io.ErrClosedPipe)
	})
}

func TestUpdateTask(t *testing.T) {
	t.Run("should update a task while verifying storage and repository interactions", func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		update := s.UpdateTaskParams{ID: 1, Description: ptrStr("cook dinner")}
		task, err := service.UpdateTask(update)

		assertNoError(t, err)
		assertTask(t, task, r.Task{})
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks, storageSaveTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryUpdateTask,
			repositoryFindAll,
		})
	})

	t.Run("should return error when storage.LoadTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.loadTasksErr = io.ErrClosedPipe
		update := s.UpdateTaskParams{ID: 1, Description: ptrStr("cook dinner")}

		_, err := service.UpdateTask(update)
		assertError(t, err, io.ErrClosedPipe)
	})

	t.Run("should return error when repository.UpdateTask fails", func(t *testing.T) {
		service, _, spyRepository := newTestService()
		spyRepository.updateTaskErr = io.ErrClosedPipe
		update := s.UpdateTaskParams{ID: 1, Description: ptrStr("cook dinner")}

		_, err := service.UpdateTask(update)
		assertError(t, err, io.ErrClosedPipe)
	})

	t.Run("should return error when storage.SaveTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.saveTasksErr = io.ErrClosedPipe
		update := s.UpdateTaskParams{ID: 1, Description: ptrStr("cook dinner")}

		_, err := service.UpdateTask(update)
		assertError(t, err, io.ErrClosedPipe)
	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("should delete a task while verifying storage and repository interactions", func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		_, err := service.DeleteTask(1)

		assertNoError(t, err)
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks, storageSaveTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryDeleteTask,
			repositoryFindAll,
		})
	})

	t.Run("should return error when storage.LoadTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.loadTasksErr = io.ErrClosedPipe

		_, err := service.DeleteTask(1)
		assertError(t, err, io.ErrClosedPipe)
	})

	t.Run("should return error when repository.DeleteTask fails", func(t *testing.T) {
		service, _, spyRepository := newTestService()
		spyRepository.deleteTaskErr = io.ErrClosedPipe

		_, err := service.DeleteTask(1)
		assertError(t, err, io.ErrClosedPipe)
	})

	t.Run("should return error when storage.SaveTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.saveTasksErr = io.ErrClosedPipe

		_, err := service.DeleteTask(1)
		assertError(t, err, io.ErrClosedPipe)
	})
}

func TestListTasks(t *testing.T) {
	t.Run("should list all tasks when nil is passed", func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		tasks, err := service.ListTasks(nil)

		assertNoError(t, err)
		assertTasks(t, tasks, r.Tasks{})
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryFindAll,
		})
	})

	t.Run(`should list all todo tasks when "todo" is passed`, func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		status := r.Status(r.StatusTodo)
		tasks, err := service.ListTasks(&status)

		assertNoError(t, err)
		assertTasks(t, tasks, r.Tasks{})
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryFindMany,
		})
	})

	t.Run(`should list all in-process tasks when "in-process" is passed`, func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		status := r.Status(r.StatusInProcess)
		tasks, err := service.ListTasks(&status)

		assertNoError(t, err)
		assertTasks(t, tasks, r.Tasks{})
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryFindMany,
		})
	})

	t.Run(`should list all done tasks when "done" is passed`, func(t *testing.T) {
		service, spyStorage, spyRepository := newTestService()
		status := r.Status(r.StatusDone)
		tasks, err := service.ListTasks(&status)

		assertNoError(t, err)
		assertTasks(t, tasks, r.Tasks{})
		assertCalls(t, spyStorage.calls, Calls{storageLoadTasks})
		assertCalls(t, spyRepository.calls, Calls{
			repositoryImportTasksData,
			repositoryFindMany,
		})
	})

	t.Run("should return an error when storage.LoadTasks fails", func(t *testing.T) {
		service, spyStorage, _ := newTestService()
		spyStorage.loadTasksErr = io.ErrClosedPipe

		_, err := service.ListTasks(nil)
		assertError(t, err, io.ErrClosedPipe)
	})
}

const (
	storageSaveTasks string = "storageSaveTasks"
	storageLoadTasks string = "storageLoadTasks"

	repositoryAddTask         string = "repositoryAddTask"
	repositoryUpdateTask      string = "repositoryUpdateTask"
	repositoryDeleteTask      string = "repositoryDeleteTask"
	repositoryFindAll         string = "repositoryFindAll"
	repositoryFindMany        string = "repositoryFindMany"
	repositoryImportTasksData string = "repositoryImportTasksData"
)

type Calls []string

type spyStorage struct {
	calls        Calls
	loadTasksErr error
	saveTasksErr error
}

func (s *spyStorage) SaveTasks(writer io.Writer, tasks r.Tasks) (r.Tasks, error) {
	s.calls = append(s.calls, storageSaveTasks)
	return r.Tasks{}, s.saveTasksErr
}

func (s *spyStorage) LoadTasks(io.Reader) (r.Tasks, error) {
	s.calls = append(s.calls, storageLoadTasks)
	return r.Tasks{}, s.loadTasksErr
}

type spyRepository struct {
	calls         Calls
	addTaskErr    error
	updateTaskErr error
	deleteTaskErr error
}

func (s *spyRepository) ImportTasksData(tasks r.Tasks) r.Tasks {
	s.calls = append(s.calls, repositoryImportTasksData)
	return r.Tasks{}
}

func (s *spyRepository) AddTask(description string) (r.Task, error) {
	s.calls = append(s.calls, repositoryAddTask)
	return r.Task{}, s.addTaskErr
}

func (s *spyRepository) UpdateTask(params r.UpdateTaskParams) (r.Task, error) {
	s.calls = append(s.calls, repositoryUpdateTask)
	return r.Task{}, s.updateTaskErr
}

func (s *spyRepository) DeleteTask(id uint) (bool, error) {
	s.calls = append(s.calls, repositoryDeleteTask)
	return true, s.deleteTaskErr
}

func (s *spyRepository) FindAll() r.Tasks {
	s.calls = append(s.calls, repositoryFindAll)
	return r.Tasks{}
}

func (s *spyRepository) FindMany(status r.Status) r.Tasks {
	s.calls = append(s.calls, repositoryFindMany)
	return r.Tasks{}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal("got an error but didn't want one")
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if got != want {
		t.Fatalf("got error %v, want %v", got, want)
	}
}

func assertCalls(t testing.TB, got, want Calls) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
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

func newTestService() (*s.TaskService, *spyStorage, *spyRepository) {
	buffer := bytes.Buffer{}
	spyStorage := &spyStorage{}
	spyRepository := &spyRepository{}
	service := &s.TaskService{
		Reader:     &buffer,
		Writer:     &buffer,
		Storage:    spyStorage,
		Repository: spyRepository,
	}

	return service, spyStorage, spyRepository
}

func ptrStr(s string) *string {
	return &s
}
