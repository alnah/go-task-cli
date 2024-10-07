package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sh "github.com/alnah/task-tracker/internal/shared"
)

func Test_InitDataError(t *testing.T) {
	err := &InitDataError{InitData: "{}"}
	errMsg := err.Error()
	sh.AssertErrorMessage(t, err, errMsg, string(err.InitData))
}

func Test_FilenameExtError(t *testing.T) {
	err := &FilenameExtErr{Filename: "test.json"}
	errMsg := err.Error()
	sh.AssertErrorMessage(t, err, errMsg, string(err.Filename))
}

func Test_JSONFileStore_InitFile_Happy(t *testing.T) {
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
			fs := JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("test_%s.json", tc.name),
				InitData: tc.initData,
			}
			filepath := filepath.Join(fs.DestDir, fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			_, err := fs.InitFile()
			sh.AssertNoError(t, err)

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("got nothing, but want a file")
			}

			content, err := os.ReadFile(filepath)
			sh.AssertNoError(t, err)

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
			sh.AssertNoError(t, err)

			var want any
			err = json.Unmarshal([]byte(tc.initData), &want)
			sh.AssertNoError(t, err)

			sh.AssertDeepEqual(t, got, want)
		})
	}
}

func Test_JSONFileStore_InitFile_Sad_Edge(t *testing.T) {
	testCases := []struct {
		name    string
		fs      JSONFileStore[any]
		errType error
	}{
		{
			"returns an InitDataError for a bad initial data structure",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "bad_init_data_struct.json",
				InitData: "incorrect",
			},
			&InitDataError{},
		},
		{
			"returns a FilenameExtError for a bad filename extension",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "bad_filename.incorrect",
				InitData: "{}",
			},
			&FilenameExtErr{},
		},
		{
			"returns an os.PathError when destinary directory creation fails",
			JSONFileStore[any]{
				DestDir:  strings.Repeat("a", 1000), // too long
				Filename: "dest_dir_creation_fails.json",
				InitData: "{}",
			},
			&os.PathError{},
		},
		{
			"returns an os.PathError when file creation fails",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("%s.json", strings.Repeat("a", 1000)), // too long
				InitData: "{}",
			},
			&os.PathError{},
		},
		{
			"returns an os.PathError for an empty destination directory",
			JSONFileStore[any]{
				DestDir:  "",
				Filename: "empty_dir.json",
				InitData: "{}",
			},
			&os.PathError{},
		},
		{
			"returns a FilenameExtErr an empty filename",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "",
				InitData: "{}",
			},
			&FilenameExtErr{},
		},
		{
			"returns an InitDataError for an empty initial data structure",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "empty_init_data.json",
				InitData: "",
			},
			&InitDataError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filepath := filepath.Join(tc.fs.DestDir, tc.fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			_, err := tc.fs.InitFile()
			assertError(t, err, tc.errType)
		})
	}
}

func Test_JSONFileStore_SaveData_Happy(t *testing.T) {
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
			sh.AssertNoError(t, err)
			t.Cleanup(func() {
				err := writeFile.Close()
				sh.AssertNoError(t, err)
			})

			var want any
			err = json.Unmarshal([]byte(tc.jsonData), &want)
			sh.AssertNoError(t, err)

			err = fs.SaveData(want, filepath)
			sh.AssertNoError(t, err)

			readFile, err := os.OpenFile(filepath, os.O_RDONLY, 0444)
			sh.AssertNoError(t, err)
			t.Cleanup(func() {
				err := readFile.Close()
				sh.AssertNoError(t, err)
			})

			bytes, err := io.ReadAll(readFile)
			sh.AssertNoError(t, err)

			var got any
			err = json.Unmarshal(bytes, &got)
			sh.AssertNoError(t, err)

			sh.AssertDeepEqual(t, got, want)
		})
	}
}

func Test_JSONFileStore_SaveData_Sad_Edge(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		data     any
		errType  error
	}{
		{
			name:     "returns an os.PathError for an empty filename",
			filename: "",
			data:     FakeJSONArray,
			errType:  &os.PathError{},
		},
		{
			name:     "returns an os.PathError for empty data",
			filename: "test_empty.json",
			data:     nil,
			errType:  &os.PathError{},
		},
		{
			name:     "returns an os.PathError if EOF",
			filename: "test_eof.json",
			data:     []byte("{\"key\": \"value\""),
			errType:  &os.PathError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fs, filepath := setupJSONFileStore(t, tc.filename)
			err := fs.SaveData(tc.data, filepath)
			assertError(t, err, tc.errType)
		})
	}
}

func Test_JSONFileStore_LoadData_Happy(t *testing.T) {
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
			sh.AssertNoError(t, err)

			got, err := fs.LoadData(filepath)
			sh.AssertNoError(t, err)

			var want any
			err = json.Unmarshal([]byte(tc.jsonData), &want)
			sh.AssertNoError(t, err)

			sh.AssertDeepEqual(t, got, want)
		})
	}
}

func Test_JSONFileStore_LoadData_Sad(t *testing.T) {
	t.Run("returns an os.PathError if the file doesn't exist", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_not_exist.json")
		got, err := fs.LoadData(filepath)
		assertError(t, err, &os.PathError{})
		sh.AssertNil(t, got)
	})

	t.Run("returns an json.SyntaxError if EOF", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_eof.json")

		err := os.WriteFile(filepath, []byte("{\"key\": \"value\""), 0644)
		sh.AssertNoError(t, err)

		got, err := fs.LoadData(filepath)
		assertError(t, err, &json.SyntaxError{})
		sh.AssertNil(t, got)
	})
}

func Test_JSONFileStore_LoadData_Edge(t *testing.T) {
	t.Run("returns an os.PathError if filepath is an empty string",
		func(t *testing.T) {
			fs, emptyFilepath := setupJSONFileStore(t, "")
			got, err := fs.LoadData(emptyFilepath)
			assertError(t, err, &os.PathError{})
			sh.AssertNil(t, got)
		})
}

func Test_JSONFileStore_Happy_closeFile(t *testing.T) {
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

func Test_JSONFileStore_Sad_closeFile(t *testing.T) {
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
	assertError(t, err, &os.PathError{})
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

func assertError(t testing.TB, err error, expectedType error) {
	t.Helper()
	sh.AssertNotNil(t, err)

	switch expectedType.(type) {
	// Custom Errors
	case *InitDataError:
		var initDataErr *InitDataError
		if !errors.As(err, &initDataErr) {
			t.Errorf("got %T, want InitDataError", err)
		}

	case *FilenameExtErr:
		var filenameErr *FilenameExtErr
		if !errors.As(err, &filenameErr) {
			t.Errorf("got %T, want FilenameError", err)
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
