package store

import "os"

func (fs *JSONFileStore[T]) openFile(filepath string, mode int) (*os.File, error) {
	file, err := os.OpenFile(filepath, mode, 0644)

	if err != nil {
		op := "opening file"
		switch mode {
		case os.O_RDONLY:
			op += " in read-only mode"
		default:
			op += " in write mode"
		}

		return nil, &StoreError{
			Operation: op,
			Message:   err.Error(),
		}
	}

	return file, nil
}
