package store

import "os"

func (fs *JSONFileStore[T]) CloseFile(file *os.File) error {
	if err := file.Close(); err != nil {
		return &StoreError{
			Operation: "closing file",
			Message:   err.Error(),
		}
	}
	return nil
}
