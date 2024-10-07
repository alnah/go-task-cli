package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Store[T any] interface {
	InitFile() (*os.File, error)
	LoadData(string) (T, error)
	SaveData(T, string) error
}

type JSONInitData string

const (
	EmptyArray  JSONInitData = "[]"
	EmptyObject JSONInitData = "{}"
)

type JSONFileStore[T any] struct {
	DestDir  string
	Filename string
	InitData JSONInitData
}

type InitDataError struct {
	InitData JSONInitData
}

func (e *InitDataError) Error() string {
	return fmt.Sprintf(
		"expected an empty array `[]` or an empty object `{}`, "+
			"but got %s", e.InitData,
	)
}

type FilenameExtErr struct {
	Filename string
}

func (e *FilenameExtErr) Error() string {
	return fmt.Sprintf("expected a `.json` file, but got `%s`", e.Filename)
}

func (fs *JSONFileStore[T]) InitFile() (*os.File, error) {
	if err := fs.validateDataStructure(); err != nil {
		return nil, err
	}

	if err := fs.validateFilenameExt(); err != nil {
		return nil, err
	}

	if err := fs.createDestDir(); err != nil {
		return nil, err
	}

	file, err := fs.createFile()
	if err != nil {
		return nil, err
	}

	defer fs.closeFile(file)

	return file, nil
}

func (fs *JSONFileStore[T]) LoadData(filepath string) (T, error) {
	var zero T

	file, err := fs.openFile(filepath, os.O_RDONLY)
	if err != nil {
		return zero, err
	}

	defer fs.closeFile(file)

	bytes, err := fs.readFileContent(file)
	if err != nil {
		return zero, err
	}

	data, err := fs.unmarshall(bytes)
	if err != nil {
		return zero, err
	}

	return data.(T), nil
}

func (fs *JSONFileStore[T]) SaveData(data T, filepath string) error {
	file, err := fs.openFile(filepath, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return err
	}

	defer fs.closeFile(file)

	bytes, err := fs.marshall(data)
	if err != nil {
		return err
	}

	if err := fs.writeFileContent(file, bytes); err != nil {
		return err
	}

	return nil
}

func (fs *JSONFileStore[T]) validateDataStructure() error {
	if fs.InitData != EmptyArray && fs.InitData != EmptyObject {
		return &InitDataError{InitData: fs.InitData}
	}

	return nil
}

func (fs *JSONFileStore[T]) validateFilenameExt() error {
	if filepath.Ext(fs.Filename) != ".json" {
		return &FilenameExtErr{Filename: fs.Filename}
	}

	return nil
}

func (fs *JSONFileStore[T]) createDestDir() error {
	if err := os.MkdirAll(fs.DestDir, 0644); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	return nil
}

func (fs *JSONFileStore[T]) createFile() (*os.File, error) {
	filepath := filepath.Join(fs.DestDir, fs.Filename)

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	file.WriteString(string(fs.InitData))

	return file, nil
}

func (fs *JSONFileStore[T]) openFile(filepath string, mode int) (*os.File, error) {
	file, err := os.OpenFile(filepath, mode, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (fs *JSONFileStore[T]) closeFile(file *os.File) error {
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

func (fs *JSONFileStore[T]) marshall(data any) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	return bytes, nil
}

func (fs *JSONFileStore[T]) unmarshall(bytes []byte) (any, error) {
	var data any

	if err := json.Unmarshal(bytes, &data); err != nil {
		var zero any
		return zero, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return data, nil
}

func (fs *JSONFileStore[T]) readFileContent(file *os.File) ([]byte, error) {
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	return bytes, nil
}

func (fs *JSONFileStore[T]) writeFileContent(file *os.File, bytes []byte) error {
	if _, err := file.Write(bytes); err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

var _ Store[any] = (*JSONFileStore[any])(nil)
