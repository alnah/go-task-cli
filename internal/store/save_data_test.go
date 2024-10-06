package store

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

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
