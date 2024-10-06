package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONFileDataStore_Happy_CloseFile(t *testing.T) {
	testCases := []struct {
		name     string
		initData JSONInitData
	}{
		{"closes a JSON file initialized with an empty array", "[]"},
		{"closes a JSON file initialized with an empty object", "{}"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := fmt.Sprintf("test_%s.json", strings.ReplaceAll(tc.name, " ", "_"))
			fs := JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: filename,
				InitData: tc.initData,
			}
			filepath := filepath.Join(fs.DestDir, fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			file, err := fs.Init()
			if err != nil {
				t.Fatalf("Init failed: %v", err)
			}

			err = fs.CloseFile(file)
			if err != nil {
				t.Fatalf("CloseFile failed: %v", err)
			}

			_, err = file.WriteString("attempt to write to a closed file")
			if err == nil {
				t.Errorf("want an error, but didn't got one, because it's closed!")
			}
		})
	}
}

func TestJSONFileDataStore_Sad_CloseFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := "test_already_closed.json"
	fs := JSONFileStore[any]{
		DestDir:  tempDir,
		Filename: filename,
		InitData: "{}",
	}
	filepath := filepath.Join(fs.DestDir, fs.Filename)
	t.Cleanup(func() { os.Remove(filepath) })

	file, err := fs.Init()
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// close the file before...
	err = file.Close()
	if err != nil {
		t.Fatalf("manual Close failed: %v", err)
	}

	// ...closing it with the CloseFile method to trigger an error
	err = fs.CloseFile(file)
	assertStoreError(t, err)
}
