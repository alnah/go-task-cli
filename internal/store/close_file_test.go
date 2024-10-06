package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJSONFileDataStore_Happy_CloseFile(t *testing.T) {
	t.Run("successfully closes a file", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_close_success.json")
		t.Cleanup(func() { os.Remove(filepath) })

		file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			t.Fatalf("file creation failed: %v", err)
		}

		err = fs.closeFile(file)
		if err != nil {
			t.Fatalf("closeFile failed: %v", err)
		}

		_, err = file.WriteString("attempt to write to a closed file")
		if err == nil {
			t.Errorf("want an error, but didn't got one")
		}
	})
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

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("File creation failed: %v", err)
	}

	err = file.Close() // close manually the file
	if err != nil {
		t.Fatalf("manual Close failed: %v", err)
	}

	err = fs.closeFile(file) // error because it has been already closed
	assertStoreError(t, err)
}
