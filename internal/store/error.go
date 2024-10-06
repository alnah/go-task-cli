package store

import "fmt"

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

