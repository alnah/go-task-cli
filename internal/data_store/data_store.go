package data_store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrCreatingDir = errors.New("error while creating dir")
var ErrCreatingFile = errors.New("error while creating file")
var ErrCheckingFile = errors.New("error while checking file")
var ErrSavingData = errors.New("error while saving data")
var ErrLoadingData = errors.New("error while loading data")

type DataStore[T any] interface {
	SaveData(T) (T, error)
	LoadData() (T, error)
}

type JSONFileDataStore[T any] struct{ Filename string }

func NewJSONFileDataStore[T any](filename string) (*JSONFileDataStore[T], error) {
	dirpath := getDirPath()
	filepath := getFilePath(filename)

	if err := os.MkdirAll(dirpath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCreatingDir, err)
	}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCreatingFile, err)
		}
		defer close(f)

		f.WriteString("{}")
	} else if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCheckingFile, err)
	}

	return &JSONFileDataStore[T]{Filename: filename}, nil
}

func (s *JSONFileDataStore[T]) SaveData(data T) (T, error) {
	filepath := getFilePath(s.Filename)
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return data, fmt.Errorf("%w > %+v\n", ErrSavingData, err)
	}

	return s.saveDataToFile(file, data)
}

func (s *JSONFileDataStore[T]) LoadData() (T, error) {
	filepath := getFilePath(s.Filename)
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0444)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("%w > %+v\n", ErrLoadingData, err)
	}

	return s.loadDataFromFile(file)
}

func (s *JSONFileDataStore[T]) SaveDataToFile(f io.ReadWriteCloser, data T) (T, error) {
	return s.saveDataToFile(f, data)
}

func (s *JSONFileDataStore[T]) LoadDataFromFile(f io.ReadWriteCloser) (T, error) {
	return s.loadDataFromFile(f)
}

func (s *JSONFileDataStore[T]) saveDataToFile(f io.ReadWriteCloser, data T) (T, error) {
	defer close(f)

	if err := json.NewEncoder(f).Encode(data); err != nil {
		return data, fmt.Errorf("%w > %+v", ErrSavingData, err)
	}

	return data, nil
}

func (s *JSONFileDataStore[T]) loadDataFromFile(f io.ReadWriteCloser) (T, error) {
	defer close(f)

	var data T
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		var zero T
		return zero, fmt.Errorf("%w > %+v\n", ErrLoadingData, err)
	}

	return data, nil
}

func getDirPath() string {
	return fmt.Sprint("../data")
}

func getFilePath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(getDirPath(), fmt.Sprintf("%s.json", filename))
}

func close(file io.ReadWriteCloser) {
	if err := file.Close(); err != nil {
		fmt.Printf("error closing file: %+v\n", err)
	}
}
