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

	s "github.com/alnah/task-tracker/internal/storage"
)

func TestSaveData(t *testing.T) {
	t.Run("should save data and return it", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove("../data")

		storage, err := s.NewJSONFileDataStore[Data](tmpfile.Name())
		assertNoError(t, err)

		savedData, err := storage.SaveData(sampleData)
		assertNoError(t, err)

		assertData(t, savedData, sampleData)
		assertFileContainsData(t, tmpfile.Name(), sampleData)
	})

	t.Run("should handle empty data gracefully", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove("../data")

		storage, err := s.NewJSONFileDataStore[Data](tmpfile.Name())
		assertNoError(t, err)

		var emptyData Data
		savedData, err := storage.SaveData(emptyData)
		assertNoError(t, err)

		assertData(t, savedData, emptyData)
		assertFileContainsData(t, tmpfile.Name(), emptyData)
	})

	t.Run("should return an error when failing to create directory", func(t *testing.T) {
		invalidPath := "/invalid/path/to/file"
		_, err := s.NewJSONFileDataStore[Data](invalidPath)
		if err == nil {
			t.Fatalf("expected error when creating JSONFileDataStore with invalid path")
		}
	})

	t.Run("should handle error during JSON encoding", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove("../data")

		badData := BadData{}

		storage, err := s.NewJSONFileDataStore[BadData](tmpfile.Name())
		assertNoError(t, err)

		_, err = storage.SaveData(badData)
		assertError(t, err, s.ErrSavingData)
	})

	t.Run("should handle error when closing file during SaveData", func(t *testing.T) {
		mockFile := &MockWriteCloser{
			File:       mustCreateTempFile(t),
			CloseError: errors.New("mock close error"),
		}
		defer os.Remove("../data")

		storage := &s.JSONFileDataStore[Data]{Filename: mockFile.Name()}
		output := captureOutput(func() {
			_, _ = storage.SaveDataToFile(mockFile, sampleData)
		})

		if !strings.Contains(output, "error closing file") {
			t.Errorf("expected error message, got %q", output)
		}
	})
}

func TestLoadData(t *testing.T) {
	t.Run("should load data", func(t *testing.T) {
		tmpfile := mustCreateTempFileWithData(t, sampleData)
		defer os.Remove("../data")

		storage, err := s.NewJSONFileDataStore[Data](tmpfile.Name())
		assertNoError(t, err)

		loadedData, err := storage.LoadData()
		assertNoError(t, err)

		assertData(t, loadedData, sampleData)
	})

	t.Run("should handle empty data", func(t *testing.T) {
		tmpfile := mustCreateTempFileWithData(t, Data{})
		defer os.Remove("../data")

		storage, err := s.NewJSONFileDataStore[Data](tmpfile.Name())
		assertNoError(t, err)

		loadedData, err := storage.LoadData()
		assertNoError(t, err)

		assertData(t, loadedData, Data{})
	})

	t.Run("should return error on invalid JSON", func(t *testing.T) {
		tmpfile := mustCreateTempFile(t)
		defer os.Remove("../data")

		if err := os.WriteFile(tmpfile.Name(), []byte(`invalid json`), 0644); err != nil {
			t.Fatalf("failed to write invalid JSON: %v", err)
		}

		storage, err := s.NewJSONFileDataStore[Data](tmpfile.Name())
		assertNoError(t, err)

		_, err = storage.LoadData()
		assertError(t, err, s.ErrLoadingData)
	})

	t.Run("should return an error when file does not exist", func(t *testing.T) {
		nonExistentFile := "/non/existent/file.json"

		_, err := s.NewJSONFileDataStore[Data](nonExistentFile)
		if err == nil {
			t.Fatalf("expected error when creating JSONFileDataStore with invalid path")
		}
	})

	t.Run("should handle error when closing file during LoadData", func(t *testing.T) {
		mockFile := &MockReadCloser{
			File:       mustCreateTempFileWithData(t, sampleData),
			CloseError: errors.New("mock close error"),
		}
		defer os.Remove("../data")

		storage := &s.JSONFileDataStore[Data]{Filename: mockFile.Name()}
		output := captureOutput(func() {
			_, _ = storage.LoadDataFromFile(mockFile)
		})

		if !strings.Contains(output, "error closing file") {
			t.Errorf("expected error message, got %q", output)
		}
	})
}

type Data struct {
	ID    uint
	Title string
}

var sampleData = Data{
	ID:    1,
	Title: "description",
}

type BadData struct{}

func (bd BadData) MarshalJSON() ([]byte, error) {
	return nil, errors.New("mock error during marshaling")
}

func (bd *BadData) UnmarshalJSON([]byte) error {
	return errors.New("mock error during unmarshaling")
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
