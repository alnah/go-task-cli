package store

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestJSONFileStorage_LoadData_Happy(t *testing.T) {
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

func TestJSONFileStorage_LoadData_Sad(t *testing.T) {
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

func TestJSONFileStorage_LoadData_Edge(t *testing.T) {
	t.Run("returns a StoreError if filepath is an empty string",
		func(t *testing.T) {
			fs, emptyFilepath := setupJSONFileStore(t, "")
			got, err := fs.LoadData(emptyFilepath)
			assertStoreError(t, err)
			assertNil(t, got)
		})
}
