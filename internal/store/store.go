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

const (
	errorHead      = "Error: "
	errorTail      = "please provide more details"
	emptyOperation = "operation is empty, "
	emptyMessage   = "message is empty, "
	emptyBoth      = "both operation and message are empty, "
)

type StoreError struct {
	Operation string
	Message   string
}

func (e *StoreError) Error() string {
	switch {
	case e.Operation == "" && e.Message == "":
		return errorHead + emptyBoth + errorTail
	case e.Operation == "":
		return errorHead + emptyOperation + errorTail
	case e.Message == "":
		return errorHead + emptyMessage + errorTail
	}

	return fmt.Sprintf("Error while %s: %s", e.Operation, e.Message)
}

func (fs *JSONFileStore[T]) InitFile() (*os.File, error) {
	if err := fs.validateDataStructure(); err != nil {
		return nil, err
	}
	if err := fs.validateFilename(); err != nil {
		return nil, err
	}
	if err := fs.createDir(); err != nil {
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

	return data, nil
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
		return &StoreError{
			Operation: "validating data structure",
			Message:   "InitDataStruct must be either '[]' or '{}'",
		}
	}
	return nil
}

func (fs *JSONFileStore[T]) validateFilename() error {
	if filepath.Ext(fs.Filename) != ".json" {
		return &StoreError{
			Operation: "validating filename",
			Message:   "Filename must have a '.json' extension",
		}
	}
	return nil
}

func (fs *JSONFileStore[T]) createDir() error {
	if err := os.MkdirAll(fs.DestDir, 0644); err != nil {
		return &StoreError{
			Operation: "creating destination directory",
			Message:   err.Error(),
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
			Message:   err.Error(),
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
			Message:   err.Error(),
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

func (fs *JSONFileStore[T]) marshall(data T) ([]byte, error) {
	bytes, err := json.Marshal(data)

	if err != nil {
		return nil, &StoreError{
			Operation: "marshalling JSON data",
			Message:   err.Error(),
		}
	}

	return bytes, nil
}

func (fs *JSONFileStore[T]) unmarshall(bytes []byte) (T, error) {
	var data T

	if err := json.Unmarshal(bytes, &data); err != nil {
		var zero T
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
