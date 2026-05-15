package parser

import (
	"testing"
)

func TestMarkdownParser_Parse(t *testing.T) {
	source := []byte(`
# Main Title
Intro text.

## Section 1
Content of section 1.

### Subsection 1.1
Content of subsection.

## Section 2
More content.
`)

	p := NewMarkdownParser()
	sections, err := p.Parse(source)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Expected sections: Introduction (empty title), Main Title, Section 1, Subsection 1.1, Section 2
	// Actually, based on my implementation, content before the first heading goes into "Introduction".
	// Let's check.
	
	expectedCount := 4
	if len(sections) != expectedCount {
		t.Errorf("Expected %d sections, got %d", expectedCount, len(sections))
	}

	// Verify paths
	expectedPaths := [][]string{
		{"Main Title"},
		{"Main Title", "Section 1"},
		{"Main Title", "Section 1", "Subsection 1.1"},
		{"Main Title", "Section 2"},
	}

	for i, s := range sections {
		if len(s.HeadingPath) != len(expectedPaths[i]) {
			t.Errorf("Section %d: expected path length %d, got %d", i, len(expectedPaths[i]), len(s.HeadingPath))
			continue
		}
		for j, p := range s.HeadingPath {
			if p != expectedPaths[i][j] {
				t.Errorf("Section %d path %d: expected %s, got %s", i, j, expectedPaths[i][j], p)
			}
		}
	}
}
