package store

import "os"

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
