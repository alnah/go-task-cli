package shared

import (
	"errors"
	"testing"
)

func Test_AssertNil(t *testing.T) {
	var got any = nil
	AssertNil(t, got)
}

func Test_AssertNoError(t *testing.T) {
	var err error = nil
	AssertNoError(t, err)
}

func Test_AssertDeepEqual(t *testing.T) {
	got := map[string]any{"key": "value"}
	want := map[string]any{"key": "value"}
	AssertDeepEqual(t, got, want)
}

func Test_AssertErrorMessage(t *testing.T) {
	err := errors.New("an error occurred")
	got := err.Error()
	want := "error"
	AssertErrorMessage(t, err, got, want)
}
