package parser

import (
	"fmt"
	"strings"

	"github.com/priya-sharma/documind/internal/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownParser parses markdown content using goldmark AST.
type MarkdownParser struct {
	md goldmark.Markdown
}

// NewMarkdownParser creates a new MarkdownParser.
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{
		md: goldmark.New(),
	}
}

// Section represents a part of a markdown document defined by a heading.
type Section struct {
	Heading     string
	HeadingLevel int
	Content     string
	StartLine   int
	EndLine     int
	HeadingPath []string
}

// Parse converts markdown bytes into a slice of semantic sections.
func (p *MarkdownParser) Parse(source []byte) ([]Section, error) {
	reader := text.NewReader(source)
	doc := p.md.Parser().Parse(reader)

	var sections []Section
	var currentSection *Section
	var headingStack []string

	// Helper to update heading stack based on level
	updateHeadingStack := func(level int, title string) {
		if level > len(headingStack) {
			headingStack = append(headingStack, title)
		} else {
			headingStack = headingStack[:level-1]
			headingStack = append(headingStack, title)
		}
	}

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			// Finish previous section
			if currentSection != nil {
				sections = append(sections, *currentSection)
			}

			level := node.Level
			title := string(node.Text(source))
			updateHeadingStack(level, title)

			// Start new section
			currentSection = &Section{
				Heading:      title,
				HeadingLevel: level,
				StartLine:    findLine(source, node.Lines().At(0).Start),
				HeadingPath:  append([]string{}, headingStack...),
			}

		case *ast.Paragraph, *ast.FencedCodeBlock, *ast.List, *ast.Blockquote:
			if currentSection == nil {
				// Content before any heading
				currentSection = &Section{
					Heading:     "Introduction",
					HeadingPath: []string{"Introduction"},
					StartLine:   1,
				}
			}
			
			// Extract content for this node
			lines := node.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				currentSection.Content += string(line.Value(source))
			}
			currentSection.Content += "\n"
			
			// Update end line
			if lines.Len() > 0 {
				currentSection.EndLine = findLine(source, lines.At(lines.Len()-1).Stop)
			}
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking AST: %w", err)
	}

	// Add last section
	if currentSection != nil {
		sections = append(sections, *currentSection)
	}

	return sections, nil
}

// findLine is a helper to find the 1-indexed line number for a byte offset.
func findLine(source []byte, offset int) int {
	line := 1
	for i := 0; i < offset && i < len(source); i++ {
		if source[i] == '\n' {
			line++
		}
	}
	return line
}

// MapToChunks converts Sections into model.Chunks.
func (p *MarkdownParser) MapToChunks(repoID, filePath string, sections []Section) []models.Chunk {
	var chunks []models.Chunk
	for _, s := range sections {
		// In a real implementation, we would use Chunker here to split large sections.
		// For now, we create one chunk per section.
		chunks = append(chunks, models.Chunk{
			RepoID:      repoID,
			FilePath:    filePath,
			Content:     s.Content,
			HeadingPath: strings.Join(s.HeadingPath, " > "),
			ChunkType:   models.ChunkTypeText, // Defaulting to text for now
			StartLine:   s.StartLine,
			EndLine:     s.EndLine,
			Metadata:    make(map[string]string),
		})
	}
	return chunks
}
