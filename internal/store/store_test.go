package store_test

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	st "github.com/alnah/task-tracker/internal/store"
	th "github.com/alnah/task-tracker/test_helpers"
)

func Test_InitDataError_Error(t *testing.T) {
	err := &st.InitDataError{InitData: "{}"}
	errMsg := err.Error()
	th.AssertErrorMessage(t, err, errMsg, string(err.InitData))
}

func Test_FilenameExtError_Error(t *testing.T) {
	err := &st.FilenameExtError{Filename: "test.json"}
	errMsg := err.Error()
	th.AssertErrorMessage(t, err, errMsg, string(err.Filename))
}

func Test_JSONFileStore_InitFile_Happy(t *testing.T) {
	testCases := []struct {
		name     string
		initData st.JSONInitData
	}{
		{"creates a JSON file with an empty array", "[]"},
		{"creates a JSON file with an empty object", "{}"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fs := st.JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("test_%s.json", tc.name),
				InitData: tc.initData,
			}
			filepath := filepath.Join(fs.DestDir, fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			_, err := fs.InitFile()
			th.AssertNoError(t, err)

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				t.Errorf("got nothing, but want a file")
			}

			content, err := os.ReadFile(filepath)
			th.AssertNoError(t, err)

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
			th.AssertNoError(t, err)

			var want any
			err = json.Unmarshal([]byte(tc.initData), &want)
			th.AssertNoError(t, err)

			th.AssertDeepEqual(t, got, want)
		})
	}
}

func Test_JSONFileStore_InitFile_Sad_Edge(t *testing.T) {
	testCases := []struct {
		name    string
		fs      st.JSONFileStore[any]
		errType error
	}{
		{
			"returns an InitDataError for a bad initial data structure",
			st.JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "bad_init_data_struct.json",
				InitData: "incorrect",
			},
			&st.InitDataError{},
		},
		{
			"returns a FilenameExtError for a bad filename extension",
			st.JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "bad_filename.incorrect",
				InitData: "{}",
			},
			&st.FilenameExtError{},
		},
		{
			"returns an os.PathError when destinary directory creation fails",
			st.JSONFileStore[any]{
				DestDir:  strings.Repeat("a", 1000), // too long
				Filename: "dest_dir_creation_fails.json",
				InitData: "{}",
			},
			&os.PathError{},
		},
		{
			"returns an os.PathError when file creation fails",
			st.JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("%s.json", strings.Repeat("a", 1000)), // too long
				InitData: "{}",
			},
			&os.PathError{},
		},
		{
			"returns an os.PathError for an empty destination directory",
			st.JSONFileStore[any]{
				DestDir:  "",
				Filename: "empty_dir.json",
				InitData: "{}",
			},
			&os.PathError{},
		},
		{
			"returns a FilenameExtErr an empty filename",
			st.JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "",
				InitData: "{}",
			},
			&st.FilenameExtError{},
		},
		{
			"returns an InitDataError for an empty initial data structure",
			st.JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "empty_init_data.json",
				InitData: "",
			},
			&st.InitDataError{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			filepath := filepath.Join(tc.fs.DestDir, tc.fs.Filename)
			t.Cleanup(func() { os.Remove(filepath) })

			_, err := tc.fs.InitFile()
			th.AssertError(t, err, tc.errType)
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
			th.AssertNoError(t, err)
			t.Cleanup(func() {
				err := writeFile.Close()
				th.AssertNoError(t, err)
			})

			var want any
			err = json.Unmarshal([]byte(tc.jsonData), &want)
			th.AssertNoError(t, err)

			err = fs.SaveData(want, filepath)
			th.AssertNoError(t, err)

			readFile, err := os.OpenFile(filepath, os.O_RDONLY, 0444)
			th.AssertNoError(t, err)
			t.Cleanup(func() {
				err := readFile.Close()
				th.AssertNoError(t, err)
			})

			bytes, err := io.ReadAll(readFile)
			th.AssertNoError(t, err)

			var got any
			err = json.Unmarshal(bytes, &got)
			th.AssertNoError(t, err)

			th.AssertDeepEqual(t, got, want)
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
			th.AssertError(t, err, tc.errType)
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
			th.AssertNoError(t, err)

			got, err := fs.LoadData(filepath)
			th.AssertNoError(t, err)

			var want any
			err = json.Unmarshal([]byte(tc.jsonData), &want)
			th.AssertNoError(t, err)

			th.AssertDeepEqual(t, got, want)
		})
	}
}

func Test_JSONFileStore_LoadData_Sad(t *testing.T) {
	t.Run("returns an os.PathError if the file doesn't exist", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_not_exist.json")
		got, err := fs.LoadData(filepath)
		th.AssertError(t, err, &os.PathError{})
		th.AssertNil(t, got)
	})

	t.Run("returns an json.SyntaxError if EOF", func(t *testing.T) {
		fs, filepath := setupJSONFileStore(t, "test_eof.json")

		err := os.WriteFile(filepath, []byte("{\"key\": \"value\""), 0644)
		th.AssertNoError(t, err)

		got, err := fs.LoadData(filepath)
		th.AssertError(t, err, &json.SyntaxError{})
		th.AssertNil(t, got)
	})
}

func Test_JSONFileStore_LoadData_Edge(t *testing.T) {
	t.Run("returns an os.PathError if filepath is an empty string",
		func(t *testing.T) {
			fs, emptyFilepath := setupJSONFileStore(t, "")
			got, err := fs.LoadData(emptyFilepath)
			th.AssertError(t, err, &os.PathError{})
			th.AssertNil(t, got)
		})
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

func setupJSONFileStore(t *testing.T, f string) (*st.JSONFileStore[any], string) {
	fs := &st.JSONFileStore[any]{
		DestDir:  t.TempDir(),
		Filename: f,
		InitData: "{}",
	}
	filepath := filepath.Join(fs.DestDir, fs.Filename)
	t.Cleanup(func() { os.Remove(filepath) })
	return fs, filepath
}
