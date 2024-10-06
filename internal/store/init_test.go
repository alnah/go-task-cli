package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
			fs := JSONFileStore[any]{
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

func TestJSONFileStore_InitFile_Sad(t *testing.T) {
	testCases := []struct {
		name string
		fs   JSONFileStore[any]
	}{
		{
			"returns a StoreError for a bad InitDataStruct",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "bad_init_data_struct.json",
				InitData: "incorrect",
			},
		},
		{
			"returns a StoreError for a bad Filename",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "bad_filename.incorrect",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError when DestDir creation fails",
			JSONFileStore[any]{
				DestDir:  strings.Repeat("a", 1000), // too long
				Filename: "dest_dir_creation_fails.json",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError when File creation fails",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: fmt.Sprintf("%s.json", strings.Repeat("a", 1000)), // too long
				InitData: "{}",
			},
		},
	}
	runTestCases(t, testCases)
}

func TestJSONFileStore_InitFile_Edge(t *testing.T) {
	testCases := []struct {
		name string
		fs   JSONFileStore[any]
	}{
		{
			"returns a StoreError for an empty DestDir",
			JSONFileStore[any]{
				DestDir:  "",
				Filename: "empty_dir.json",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError for an empty Filename",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "",
				InitData: "{}",
			},
		},
		{
			"returns a StoreError for an empty InitData",
			JSONFileStore[any]{
				DestDir:  t.TempDir(),
				Filename: "empty_init_data.json",
				InitData: "",
			},
		},
	}
	runTestCases(t, testCases)
}

func runTestCases(t *testing.T, testCases []struct {
	name string
	fs   JSONFileStore[any]
}) {
	t.Helper()
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
