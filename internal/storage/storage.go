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

type Storage[T any] interface {
	SaveData(T) (T, error)
	LoadData() (T, error)
}

type JSONStorage[T any] struct{ Filepath string }

func (s *JSONStorage[T]) SaveData(data T) (T, error) {
	file, err := os.OpenFile(s.Filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return data, fmt.Errorf("%w: %+v", ErrSavingData, err)
	}

	return s.saveDataToFile(file, data)
}

func (s *JSONStorage[T]) LoadData() (T, error) {
	file, err := os.OpenFile(s.Filepath, os.O_RDONLY, 0444)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%w: %+v", ErrLoadingData, err)
	}

	return s.loadDataFromFile(file)
}

func (s *JSONStorage[T]) SaveDataToFile(file io.WriteCloser, data T) (T, error) {
	return s.saveDataToFile(file, data)
}

func (s *JSONStorage[T]) LoadDataFromFile(file io.ReadCloser) (T, error) {
	return s.loadDataFromFile(file)
}

func (s *JSONStorage[T]) saveDataToFile(file io.WriteCloser, data T) (T, error) {
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing file: %+v\n", err)
		}
	}()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return data, fmt.Errorf("%w: %+v", ErrSavingData, err)
	}

	return data, nil
}

func (s *JSONStorage[T]) loadDataFromFile(file io.ReadCloser) (T, error) {
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing file: %+v\n", err)
		}
	}()

	var data T
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		var zero T
		return zero, fmt.Errorf("%w: %+v", ErrLoadingData, err)
	}

	return data, nil
}
