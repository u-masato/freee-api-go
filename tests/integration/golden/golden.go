// Package golden provides utilities for golden file testing.
//
// Golden files store expected test outputs and are used to verify
// that API responses match expected formats. This is useful for:
//   - Verifying response structures haven't changed unexpectedly
//   - Documenting expected API response formats
//   - Regression testing
//
// Usage:
//
//	// In your test
//	golden.Assert(t, "testname", actualData)
//
//	// To update golden files (run with -update flag)
//	// go test ./tests/integration/... -update
package golden

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// Assert compares actual data with the golden file.
// If the -update flag is set, it updates the golden file instead.
func Assert(t *testing.T, name string, actual interface{}) {
	t.Helper()

	// Marshal actual data to JSON
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal actual data: %v", err)
	}

	goldenPath := filepath.Join("testdata", name+".golden.json")

	if *update {
		// Update the golden file
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("Failed to create testdata directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, actualJSON, 0644); err != nil {
			t.Fatalf("Failed to update golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read the golden file
	expectedJSON, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file does not exist: %s (run with -update to create)", goldenPath)
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	// Normalize both by trimming whitespace and comparing
	actualNormalized := bytes.TrimSpace(actualJSON)
	expectedNormalized := bytes.TrimSpace(expectedJSON)

	// Compare
	if !bytes.Equal(actualNormalized, expectedNormalized) {
		t.Errorf("Golden file mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s", name, expectedNormalized, actualNormalized)
	}
}

// AssertJSON compares actual JSON bytes with the golden file.
func AssertJSON(t *testing.T, name string, actual []byte) {
	t.Helper()

	// Pretty-print the JSON
	var data interface{}
	if err := json.Unmarshal(actual, &data); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	Assert(t, name, data)
}

// Load loads a golden file and returns its contents.
func Load(t *testing.T, name string) []byte {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden.json")
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Golden file does not exist: %s", goldenPath)
		}
		t.Fatalf("Failed to read golden file: %v", err)
	}

	return data
}

// LoadJSON loads a golden file and unmarshals it into the provided value.
func LoadJSON(t *testing.T, name string, v interface{}) {
	t.Helper()

	data := Load(t, name)
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("Failed to unmarshal golden file: %v", err)
	}
}
