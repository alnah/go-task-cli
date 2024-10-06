package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDataStore_Error_Happy(t *testing.T) {
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

func TestDataStore_Error_Sad_Edge(t *testing.T) {
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

func TestJSONFileStore_InitFile_Happy(t *testing.T) {
	testCases := []struct {
		name     string
		initData JSONInitData
	}{
		{"creates a JSON file with an empty array", "[]"},
		{"creates a JSON file with an empty object", "{}"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fs := JSONFileStore[any, any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("test_%s.json", tc.name),
				InitData: tc.initData,
			}
			filepath := filepath.Join(fs.DestDir, fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			_, err := fs.InitFile()
			assertNoError(t, err)

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("got nothing, but want a file")
			}

			content, err := os.ReadFile(filepath)
			assertNoError(t, err)

			var got any
			switch tc.initData {
			case "[]":
				var gotArray []any
				err = json.Unmarshal(content, &gotArray)
				got = gotArray
			case "{}":
				var gotObject map[string]any
				err = json.Unmarshal(content, &gotObject)
				got = gotObject
			}
			assertNoError(t, err)

			var want any
			err = json.Unmarshal([]byte(tc.initData), &want)
			assertNoError(t, err)

			assertDeepEqual(t, got, want)
		})
	}
}

func TestJSONFileStore_InitFile_Sad_Edge(t *testing.T) {
	testCases := []struct {
		name string
		fs   JSONFileStore[any, any]
	}{
		{
			"returns a StoreError for a bad InitDataStruct",
			JSONFileStore[any, any]{
				DestDir:  t.TempDir(),
				Filename: "bad_init_data_struct.json",
				InitData: "incorrect",
			},
		},
		{
			"returns a StoreError for a bad Filename",
			JSONFileStore[any, any]{
				DestDir:  t.TempDir(),
				Filename: "bad_filename.incorrect",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError when DestDir creation fails",
			JSONFileStore[any, any]{
				DestDir:  strings.Repeat("a", 1000), // too long
				Filename: "dest_dir_creation_fails.json",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError when File creation fails",
			JSONFileStore[any, any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("%s.json", strings.Repeat("a", 1000)), // too long
				InitData: "{}",
			},
		},
		{
			"returns a StoreError for an empty DestDir",
			JSONFileStore[any, any]{
				DestDir:  "",
				Filename: "empty_dir.json",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError for an empty Filename",
			JSONFileStore[any, any]{
				DestDir:  t.TempDir(),
				Filename: "",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError for an empty InitData",
			JSONFileStore[any, any]{
				DestDir:  t.TempDir(),
				Filename: "empty_init_data.json",
				InitData: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filepath := filepath.Join(tc.fs.DestDir, tc.fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			_, err := tc.fs.InitFile()
			assertStoreError(t, err)
		})
	}
}

func TestJSONFileStore_SaveData_Happy(t *testing.T) {
	testCases := []struct {
		name     string
		jsonData string
	}{
		{"loads JSONArray", FakeJSONArray},
		{"loads JSONObject", FakeJSONObject},
		{"loads NestedJSONArray", FakeNestedJSONArray},
		{"loads NestedJSONObject", FakeNestedJSONObject},
		{"loads NestedJSONMixed", FakeNestedMixedData},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filename := strings.ReplaceAll(tc.name, " ", "_")
			fs, filepath := setupJSONFileStore(t, filename)

			flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			writeFile, err := os.OpenFile(filepath, flags, 0644)
			assertNoError(t, err)
			t.Cleanup(func() {
				err := writeFile.Close()
				assertNoError(t, err)
			})

			var want any
			err = json.Unmarshal([]byte(tc.jsonData), &want)
			assertNoError(t, err)

			err = fs.SaveData(want, filepath)
			assertNoError(t, err)

			readFile, err := os.OpenFile(filepath, os.O_RDONLY, 0444)
			assertNoError(t, err)
			t.Cleanup(func() {
				err := readFile.Close()
				assertNoError(t, err)
			})

			bytes, err := io.ReadAll(readFile)
			assertNoError(t, err)

			var got any
			err = json.Unmarshal(bytes, &got)
			assertNoError(t, err)

			assertDeepEqual(t, got, want)
		})
	}
}

func TestJSONFileStore_SaveData_Sad_Edge(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		data     any
	}{
		{"empty filename", "", FakeJSONArray},
		{"empty data", "test_empty.json", nil},
		{"EOF", "test_eof.json", []byte("{\"key\": \"value\"")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, filepath := setupJSONFileStore(t, tc.filename)
			err := fs.SaveData(tc.data, filepath)
			assertStoreError(t, err)
		})
	}
}

func TestJSONFileStore_LoadData_Happy(t *testing.T) {
	testCases := []struct {
		name     string
		jsonData string
	}{
		{"loads JSONArray", FakeJSONArray},
		{"loads JSONObject", FakeJSONObject},
		{"loads NestedJSONArray", FakeNestedJSONArray},
		{"loads NestedJSONObject", FakeNestedJSONObject},
		{"loads NestedJSONMixed", FakeNestedMixedData},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filename := strings.ReplaceAll(tc.name, " ", "_") + ".json"
			fs, filepath := setupJSONFileStore(t, filename)

			err := os.WriteFile(filepath, []byte(tc.jsonData), 0644)
			assertNoError(t, err)

			got, err := fs.LoadData(filepath)
			assertNoError(t, err)

			var want any
			err = json.Unmarshal([]byte(tc.jsonData), &want)
			assertNoError(t, err)

			assertDeepEqual(t, got, want)
		})
	}
}

func TestJSONFileStore_LoadData_Sad(t *testing.T) {
	t.Run("returns a StoreError if the file doesn't exist", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_not_exist.json")
		got, err := fs.LoadData(filepath)
		assertStoreError(t, err)
		assertNil(t, got)
	})

	t.Run("returns a StoreError if EOF", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_eof.json")

		err := os.WriteFile(filepath, []byte("{\"key\": \"value\""), 0644)
		assertNoError(t, err)

		got, err := fs.LoadData(filepath)
		assertStoreError(t, err)
		assertNil(t, got)
	})
}

func TestJSONFileStore_LoadData_Edge(t *testing.T) {
	t.Run("returns a StoreError if filepath is an empty string",
		func(t *testing.T) {
			fs, emptyFilepath := setupJSONFileStore(t, "")
			got, err := fs.LoadData(emptyFilepath)
			assertStoreError(t, err)
			assertNil(t, got)
		})
}

func TestJSONFileStore_Happy_closeFile(t *testing.T) {
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

func TestJSONFileStore_Sad_closeFile(t *testing.T) {
	tempDir := t.TempDir()
	filename := "test_already_closed.json"
	fs := JSONFileStore[any, any]{
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

const (
	FakeJSONArray = `[
		{"id": 1, "name": "Item 1"},
		{"id": 2, "name": "Item 2"},
		{"id": 3, "name": "Item 3"}
	]`

	FakeJSONObject = `{
		"items": [
			{"id": 1, "name": "Item 1"},
			{"id": 2, "name": "Item 2"}
		],
		"meta": {
			"total": 2,
			"page": 1
		}
	}`

	FakeNestedJSONArray = `[
		{
			"id": 1,
			"name": "Item 1",
			"tags": ["tag1", "tag2"]
		},
		{
			"id": 2,
			"name": "Item 2",
			"tags": ["tag3", "tag4"]
		}
	]`

	FakeNestedJSONObject = `{
		"users": [
			{
				"id": 1,
				"name": "User 1",
				"address": {
					"street": "123 Main St",
					"city": "Anytown"
				}
			},
			{
				"id": 2,
				"name": "User 2",
				"address": {
					"street": "456 Elm St",
					"city": "Othertown"
				}
			}
		]
	}`

	FakeNestedMixedData = `{
		"products": [
			{"id": 1, "name": "Product 1", "categories": ["cat1", "cat2"]},
			{"id": 2, "name": "Product 2", "categories": ["cat3", "cat4"]}
		],
		"orders": {
			"orderId": 123,
			"items": [
				{"productId": 1, "quantity": 2},
				{"productId": 2, "quantity": 1}
			]
		}
	}`
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

func setupJSONFileStore(t *testing.T, f string) (*JSONFileStore[any, any], string) {
	fs := &JSONFileStore[any, any]{
		DestDir:  t.TempDir(),
		Filename: f,
		InitData: "{}",
	}
	filepath := filepath.Join(fs.DestDir, fs.Filename)
	t.Cleanup(func() { os.Remove(filepath) })
	return fs, filepath
}
