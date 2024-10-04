package storage_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	r "github.com/alnah/task-tracker/internal/repository"
	s "github.com/alnah/task-tracker/internal/storage"
)

func TestSaveData(t *testing.T) {
	t.Run("should save data and return it", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove(tmpfile.Name())

		storage := &s.JSONStorage[r.Tasks]{Filepath: tmpfile.Name()}
		savedData, err := storage.SaveData(sampleTasks)

		assertNoError(t, err)
		assertData(t, savedData, sampleTasks)
		assertFileContainsData(t, tmpfile.Name(), sampleTasks)
	})

	t.Run("should handle empty data gracefully", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove(tmpfile.Name())

		storage := &s.JSONStorage[r.Tasks]{Filepath: tmpfile.Name()}
		var emptyTasks r.Tasks
		savedData, err := storage.SaveData(emptyTasks)

		assertNoError(t, err)
		assertData(t, savedData, emptyTasks)
		assertFileContainsData(t, tmpfile.Name(), emptyTasks)
	})

	t.Run("should return an error when failing to save data", func(t *testing.T) {
		storage := &s.JSONStorage[r.Tasks]{Filepath: "/invalid/path/to/file.json"}
		_, err := storage.SaveData(sampleTasks)

		assertError(t, err, s.ErrSavingData)
	})

	t.Run("should handle error during JSON encoding", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove(tmpfile.Name())

		// Simulate JSON encoding error
		badData := BadData{}

		storage := &s.JSONStorage[BadData]{Filepath: tmpfile.Name()}
		_, err := storage.SaveData(badData)

		assertError(t, err, s.ErrSavingData)
	})

	t.Run("should handle error when closing file during SaveData", func(t *testing.T) {
		mockFile := &MockWriteCloser{
			File:       mustCreateTempFile(t),
			CloseError: errors.New("mock close error"),
		}
		defer os.Remove(mockFile.Name())

		storage := &s.JSONStorage[r.Tasks]{Filepath: mockFile.Name()}
		output := captureOutput(func() {
			_, _ = storage.SaveDataToFile(mockFile, sampleTasks)
		})

		if !strings.Contains(output, "error closing file") {
			t.Errorf("expected error message, got %q", output)
		}
	})
}

func TestLoadData(t *testing.T) {
	t.Run("should load data", func(t *testing.T) {
		tmpfile := mustCreateTempFileWithData(t, sampleTasks)
		defer os.Remove(tmpfile.Name())

		storage := &s.JSONStorage[r.Tasks]{Filepath: tmpfile.Name()}
		loadedData, err := storage.LoadData()

		assertNoError(t, err)
		assertData(t, loadedData, sampleTasks)
	})

	t.Run("should handle empty data", func(t *testing.T) {
		tmpfile := mustCreateTempFileWithData(t, r.Tasks{})
		defer os.Remove(tmpfile.Name())

		storage := &s.JSONStorage[r.Tasks]{Filepath: tmpfile.Name()}
		loadedData, err := storage.LoadData()

		assertNoError(t, err)
		assertData(t, loadedData, r.Tasks{})
	})

	t.Run("should return error on invalid JSON", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove(tmpfile.Name())

		// Write invalid JSON to the file
		if err := os.WriteFile(tmpfile.Name(), []byte(`invalid json`), 0644); err != nil {
			t.Fatalf("failed to write invalid JSON: %v", err)
		}

		storage := &s.JSONStorage[r.Tasks]{Filepath: tmpfile.Name()}
		_, err := storage.LoadData()

		assertError(t, err, s.ErrLoadingData)
	})

	t.Run("should return an error when failing to load data", func(t *testing.T) {
		storage := &s.JSONStorage[r.Tasks]{Filepath: "/non/existent/file.json"}
		_, err := storage.LoadData()

		assertError(t, err, s.ErrLoadingData)
	})

	t.Run("should handle error when closing file during LoadData", func(t *testing.T) {
		mockFile := &MockReadCloser{
			File:       mustCreateTempFileWithData(t, sampleTasks),
			CloseError: errors.New("mock close error"),
		}
		defer os.Remove(mockFile.Name())

		storage := &s.JSONStorage[r.Tasks]{Filepath: mockFile.Name()}
		output := captureOutput(func() {
			_, _ = storage.LoadDataFromFile(mockFile)
		})

		if !strings.Contains(output, "error closing file") {
			t.Errorf("expected error message, got %q", output)
		}
	})
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

type BadData struct{}

func (bd BadData) MarshalJSON() ([]byte, error) {
	return nil, errors.New("mock error during marshaling")
}

func (bd *BadData) UnmarshalJSON([]byte) error {
	return nil
}

type MockWriteCloser struct {
	*os.File
	CloseError error
}

func (mwc *MockWriteCloser) Close() error {
	if mwc.CloseError != nil {
		return mwc.CloseError
	}
	return mwc.File.Close()
}

type MockReadCloser struct {
	*os.File
	CloseError error
}

func (mrc *MockReadCloser) Close() error {
	if mrc.CloseError != nil {
		return mrc.CloseError
	}
	return mrc.File.Close()
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old
	io.Copy(&buf, r)
	return buf.String()
}

func mustCreateTempFile(t *testing.T) *os.File {
	t.Helper()
	tmpfile, err := os.CreateTemp("", "test_data_*.json")
	if err != nil {
		t.Fatal(err)
	}
	return tmpfile
}

func mustCreateTempFileWithData[T any](t *testing.T, data T) *os.File {
	t.Helper()
	tmpfile := mustCreateTempFile(t)
	writeDataToFile(t, tmpfile.Name(), data)
	return tmpfile
}

func writeDataToFile[T any](t *testing.T, filepath string, data T) {
	t.Helper()
	bytes, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal data: %v", err)
	}
	if err := os.WriteFile(filepath, bytes, 0644); err != nil {
		t.Fatalf("failed to write data to file: %v", err)
	}
}

func assertFileContainsData[T any](t *testing.T, filepath string, expectedData T) {
	t.Helper()
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var dataFromFile T
	if err := json.Unmarshal(bytes, &dataFromFile); err != nil {
		t.Fatalf("failed to unmarshal data from file: %v", err)
	}

	if !reflect.DeepEqual(dataFromFile, expectedData) {
		t.Fatalf("data in file does not match expected data")
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got an error but didn't want one: %v", err)
	}
}

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("got error %v, want error %v", got, want)
	}
}

func assertData[T any](t testing.TB, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
