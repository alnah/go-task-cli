package store

import (
	"encoding/json"
	"os"
)

func (fs *JSONFileStore[T]) SaveData(data T, filepath string) error {
	file, err := fs.openFile(filepath, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return err
	}
	defer fs.CloseFile(file)

	bytes, err := fs.marshall(data)
	if err != nil {
		return err
	}

	if err := fs.writeFileContent(file, bytes); err != nil {
		return err
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

func (fs *JSONFileStore[T]) writeFileContent(file *os.File, bytes []byte) error {
	if _, err := file.Write(bytes); err != nil {
		return &StoreError{
			Operation: "writing to file",
			Message:   err.Error(),
		}
	}

	return nil
}
