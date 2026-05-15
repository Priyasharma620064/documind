package ingestion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalker_Walk(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "documind-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	filesToCreate := map[string]string{
		"README.md":          "# Test Repo",
		"docs/intro.md":      "Introduction content",
		"src/main.go":        "package main",
		"temp.txt":           "skip this",
		".git/config":        "git data",
		"vendor/dep/file.go": "vendor data",
	}

	for path, content := range filesToCreate {
		fullPath := filepath.Join(tempDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}

	// Run walker
	walker := NewWalker([]string{".md", ".go"}, 1)
	files, err := walker.Walk(tempDir)
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Verify results
	// Expected files: README.md, docs/intro.md, src/main.go
	expectedCount := 3
	if len(files) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(files))
	}

	foundPaths := make(map[string]bool)
	for _, f := range files {
		foundPaths[f.Path] = true
	}

	expectedPaths := []string{"README.md", "docs/intro.md", "src/main.go"}
	for _, path := range expectedPaths {
		if !foundPaths[path] {
			t.Errorf("Expected path %s not found", path)
		}
	}

	// Verify .git and vendor were skipped
	if foundPaths[".git/config"] || foundPaths["vendor/dep/file.go"] {
		t.Errorf("Hidden or vendor directories were not skipped")
	}
}
