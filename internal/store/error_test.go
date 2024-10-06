package store

import (
	"testing"
)

func TestDataStoreError_Happy(t *testing.T) {
	err := &StoreError{
		Operation: "opening file",
		Message:   "file not found",
	}

	got := err.Error()
	want := "Error while opening file: file not found"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestDataStoreError_Edge(t *testing.T) {
	testCases := []struct {
		name      string
		operation string
		message   string
		want      string
	}{
		{
			name:      "empty operation and empty message",
			operation: "",
			message:   "",
			want: "Error: both operation and message are empty, " +
				"please provide more details",
		},
		{
			name:      "non-empty operation with empty message",
			operation: "some operation",
			message:   "",
			want:      "Error: message is empty, please provide more details",
		},
		{
			name:      "empty operation with non-empty message",
			operation: "",
			message:   "some message",
			want:      "Error: operation is empty, please provide more details",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := &StoreError{
				Operation: tc.operation,
				Message:   tc.message,
			}

			got := err.Error()
			if got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}
