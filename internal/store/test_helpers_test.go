package store

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func assertNil(t *testing.T, got any) {
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func assertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got an error but didn't want one:\n%v", err)
	}
}

func assertDeepEqual(t testing.TB, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func assertStoreError(t testing.TB, err error) {
	t.Helper()
	if _, ok := err.(*StoreError); !ok {
		t.Errorf("got %T, want %T", err, &StoreError{})
	}
}

func setupJSONFileStore(t *testing.T, f string) (*JSONFileStore[any], string) {
	fs := &JSONFileStore[any]{
		DestDir:  t.TempDir(),
		Filename: f,
		InitData: "{}",
	}
	filepath := filepath.Join(fs.DestDir, fs.Filename)
	t.Cleanup(func() { os.Remove(filepath) })
	return fs, filepath
}
