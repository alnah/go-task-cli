package test_helpers

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	st "github.com/alnah/task-tracker/internal/store"
	tk "github.com/alnah/task-tracker/internal/task"
)

func AssertNil(t testing.TB, got any) {
	t.Helper()
	if got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func AssertNotNil(t testing.TB, got any) {
	t.Helper()
	if got == nil {
		t.Fatalf("got nil, want non-nil value")
	}
}

func AssertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got an error, didn't want one: %v", err)
	}
}

func AssertErrorMessage(t testing.TB, err error, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want a message containing %s", got, want)
	}
}

func AssertDeepEqual(t testing.TB, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func AssertError(t testing.TB, err error, expectedType error) {
	t.Helper()
	AssertNotNil(t, err)

	switch expectedType.(type) {
	// Custom Errors
	case *st.InitDataError:
		var initDataErr *st.InitDataError
		if !errors.As(err, &initDataErr) {
			t.Errorf("got %T, want InitDataError", err)
		}

	case *st.FilenameExtError:
		var filenameErr *st.FilenameExtError
		if !errors.As(err, &filenameErr) {
			t.Errorf("got %T, want FilenameError", err)
		}

	case *tk.DescriptionError:
		var initDataErr *tk.DescriptionError
		if !errors.As(err, &initDataErr) {
			t.Errorf("got %T, want DescriptionError", err)
		}

	case *tk.TaskNotFoundError:
		var notFoundErr *tk.TaskNotFoundError
		if !errors.As(err, &notFoundErr) {
			t.Errorf("got %T, want TaskNotFoundError", err)
		}

	// Go Errors
	case *os.PathError:
		var pathErr *os.PathError
		if !errors.As(err, &pathErr) {
			t.Errorf("got %T, want os.PathError", err)
		}

	case *json.SyntaxError:
		var syntaxErr *json.SyntaxError
		if !errors.As(err, &syntaxErr) {
			t.Errorf("got %T, want json.SyntaxError", err)
		}

	default:
		t.Fatalf("got unexpected error type: %T", err)
	}
}
