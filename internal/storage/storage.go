package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

var ErrSavingData = errors.New("failed to save data")
var ErrLoadingData = errors.New("failed to load data")

type DataStore[T any] interface {
	SaveData(T) (T, error)
	LoadData() (T, error)
}

type JSONFileDataStore[T any] struct{ Filepath string }

func NewJSONFileDataStore[T any](filepath string) *JSONFileDataStore[T] {
	return &JSONFileDataStore[T]{Filepath: filepath}
}

func (s *JSONFileDataStore[T]) SaveData(data T) (T, error) {
	file, err := os.OpenFile(s.Filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return data, fmt.Errorf("%w: %+v", ErrSavingData, err)
	}

	return s.saveDataToFile(file, data)
}

func (s *JSONFileDataStore[T]) LoadData() (T, error) {
	file, err := os.OpenFile(s.Filepath, os.O_RDONLY, 0444)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%w: %+v", ErrLoadingData, err)
	}

	return s.loadDataFromFile(file)
}

func (s *JSONFileDataStore[T]) SaveDataToFile(f io.WriteCloser, data T) (T, error) {
	return s.saveDataToFile(f, data)
}

func (s *JSONFileDataStore[T]) LoadDataFromFile(f io.ReadCloser) (T, error) {
	return s.loadDataFromFile(f)
}

func (s *JSONFileDataStore[T]) saveDataToFile(f io.WriteCloser, data T) (T, error) {
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing file: %+v\n", err)
		}
	}()

	if err := json.NewEncoder(f).Encode(data); err != nil {
		return data, fmt.Errorf("%w: %+v", ErrSavingData, err)
	}

	return data, nil
}

func (s *JSONFileDataStore[T]) loadDataFromFile(f io.ReadCloser) (T, error) {
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("error closing file: %+v\n", err)
		}
	}()

	var data T
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		var zero T
		return zero, fmt.Errorf("%w: %+v", ErrLoadingData, err)
	}

	return data, nil
}
