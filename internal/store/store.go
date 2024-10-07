package store

import (
	"encoding/json"
	"errors"
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

var BadInitDataStruct = errors.New("InitDataStruct must be either '[]' or '{}'")
var BadFilenameExtension = errors.New("Filename must have a '.json' extension")

type StoreError struct {
	Operation string
	Err       error
}

func (e *StoreError) Error() string {
	return fmt.Sprintf("Store Error while %s\n>\t%s", e.Operation, e.Err)
}

func (e *StoreError) Unwrap() error {
	return e.Err
}

func (fs *JSONFileStore[T]) InitFile() (*os.File, error) {
	if err := fs.validateDataStructure(); err != nil {
		return nil, err
	}
	if err := fs.validateFilename(); err != nil {
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

func (fs *JSONFileStore[T]) LoadData(filepath string) (any, error) {
	var zero any

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

	return data, nil
}

func (fs *JSONFileStore[T]) SaveData(data any, filepath string) error {
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
		return &StoreError{
			Operation: "validating data structure",
			Err:       BadInitDataStruct,
		}
	}
	return nil
}

func (fs *JSONFileStore[T]) validateFilename() error {
	if filepath.Ext(fs.Filename) != ".json" {
		return &StoreError{
			Operation: "validating filename",
			Err:       BadFilenameExtension,
		}
	}
	return nil
}

func (fs *JSONFileStore[T]) createDestDir() error {
	if err := os.MkdirAll(fs.DestDir, 0644); err != nil {
		return &StoreError{
			Operation: "creating destination directory",
			Err:       err,
		}
	}
	return nil
}

func (fs *JSONFileStore[T]) createFile() (*os.File, error) {
	filepath := filepath.Join(fs.DestDir, fs.Filename)
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, &StoreError{
			Operation: "creating file",
			Err:       err,
		}
	}
	file.WriteString(string(fs.InitData))
	return file, nil
}

func (fs *JSONFileStore[T]) openFile(filepath string, mode int) (*os.File, error) {
	file, err := os.OpenFile(filepath, mode, 0644)

	if err != nil {
		operation := "opening file"
		switch mode {
		case os.O_RDONLY:
			operation += " in read-only mode"
		default:
			operation += " in write mode"
		}

		return nil, &StoreError{
			Operation: operation,
			Err:       err.(),
		}
	}

	return file, nil
}

func (fs *JSONFileStore[T]) closeFile(file *os.File) error {
	if err := file.Close(); err != nil {
		return &StoreError{
			Operation: "closing file",
			Message:   err.Error(),
		}
	}
	return nil
}

func (fs *JSONFileStore[T]) marshall(data any) ([]byte, error) {
	bytes, err := json.Marshal(data)

	if err != nil {
		return nil, &StoreError{
			Operation: "marshalling JSON data",
			Message:   err.Error(),
		}
	}

	return bytes, nil
}

func (fs *JSONFileStore[T]) unmarshall(bytes []byte) (any, error) {
	var data any

	if err := json.Unmarshal(bytes, &data); err != nil {
		var zero any
		return zero, &StoreError{
			Operation: "unmarshalling JSON data",
			Message:   err.Error(),
		}
	}

	return data, nil
}

func (fs *JSONFileStore[T]) readFileContent(file *os.File) ([]byte, error) {
	bytes, err := io.ReadAll(file)

	if err != nil {
		return nil, &StoreError{
			Operation: "reading file content",
			Message:   err.Error(),
		}
	}

	return bytes, nil
}

func (fs *JSONFileStore[T]) writeFileContent(file *os.File, bytes []byte) error {
	if _, err := file.Write(bytes); err != nil {
		return &StoreError{
			Operation: "writing to file",
			Message:   err.Error(),
		}
	}

	return nil
}

var _ Store[any] = (*JSONFileStore[any])(nil)
