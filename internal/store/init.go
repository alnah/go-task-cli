package store

import (
	"os"
	"path/filepath"
)

func (fs *JSONFileStore[T]) Init() (*os.File, error) {
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
	return file, nil
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

func (fs *JSONFileStore[T]) createDestDir() error {
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
