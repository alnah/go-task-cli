package storage_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	r "github.com/alnah/task-tracker/internal/repository"
	s "github.com/alnah/task-tracker/internal/storage"
)

func TestSaveTasks(t *testing.T) {
	t.Run("should save tasks and return them", func(t *testing.T) {
		storage := &s.TaskStorage{}
		buffer := &bytes.Buffer{}
		savedTasks, err := storage.SaveTasks(buffer, sampleTasks)

		assertNoError(t, err)
		assertTasks(t, savedTasks, sampleTasks)
		assertBufferContainsTasks(t, buffer, sampleTasks)
	})

	t.Run("should handle empty tasks gracefully", func(t *testing.T) {
		storage := &s.TaskStorage{}
		buffer := &bytes.Buffer{}
		savedTasks, err := storage.SaveTasks(buffer, r.Tasks{})

		assertNoError(t, err)
		assertTasks(t, savedTasks, r.Tasks{})
		assertBufferContainsTasks(t, buffer, r.Tasks{})
	})

	t.Run("should return an error when failing to save tasks", func(t *testing.T) {
		storage := &s.TaskStorage{}
		writer := &errorWriter{}
		_, err := storage.SaveTasks(writer, sampleTasks)

		assertError(t, err, s.SavingTasksError)
	})
}

func TestLoadTasks(t *testing.T) {
	t.Run("should load tasks", func(t *testing.T) {
		storage := &s.TaskStorage{}
		data, _ := json.Marshal(sampleTasks)
		buffer := bytes.NewBuffer(data)
		loadedTasks, err := storage.LoadTasks(buffer)

		assertNoError(t, err)
		assertTasks(t, loadedTasks, sampleTasks)
	})

	t.Run("should handle empty tasks", func(t *testing.T) {
		storage := &s.TaskStorage{}
		buffer := bytes.NewBuffer([]byte(`[]`))
		loadedTasks, err := storage.LoadTasks(buffer)

		assertError(t, err, s.LoadingTasksError)
		assertTasks(t, loadedTasks, r.Tasks{})
	})

	t.Run("should return error on invalid JSON", func(t *testing.T) {
		storage := &s.TaskStorage{}
		buffer := bytes.NewBuffer([]byte(`[invalid json]`))
		_, err := storage.LoadTasks(buffer)

		assertError(t, err, s.LoadingTasksError)
	})

	t.Run("should return an error when failing to load tasks", func(t *testing.T) {
		storage := &s.TaskStorage{}
		reader := &errorReader{}
		_, err := storage.LoadTasks(reader)

		assertError(t, err, s.LoadingTasksError)
	})
}

var _ s.Storage = (*s.TaskStorage)(nil)

type errorWriter struct{}

type errorReader struct{}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
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

func assertString(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func assertTasks(t testing.TB, got, want r.Tasks) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func assertBufferContainsTasks(t testing.TB, buffer *bytes.Buffer, tasks r.Tasks) {
	t.Helper()
	expectedBuffer := &bytes.Buffer{}
	if err := json.NewEncoder(expectedBuffer).Encode(tasks); err != nil {
		t.Fatalf("failed to encode tasks: %v", err)
	}
	if buffer.String() != expectedBuffer.String() {
		t.Fatalf("got %q, want %q", buffer.String(), expectedBuffer.String())
	}
}

var fixedTime = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)

var sampleTasks = r.Tasks{
	1: {
		ID:          1,
		Description: "description",
		Status:      r.StatusTodo,
		CreatedAt:   fixedTime,
		UpdatedAt:   fixedTime,
	},
}
