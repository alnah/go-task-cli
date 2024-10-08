package test_helpers

import (
	"reflect"
	"strings"
	"testing"
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
