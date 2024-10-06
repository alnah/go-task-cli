package store

import (
	"encoding/json"
	"io"
	"os"
)

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
