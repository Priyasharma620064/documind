package parser

import (
	"fmt"
	"strings"

	"github.com/priya-sharma/documind/internal/models"
)

// Chunker handles splitting large text sections into manageable semantic chunks.
type Chunker struct {
	MaxChunkSize int // Max characters per chunk
	Overlap      int // Overlap between chunks
}

// NewChunker creates a new Chunker with specified limits.
func NewChunker(maxSize, overlap int) *Chunker {
	if maxSize <= 0 {
		maxSize = 1000 // Default to 1000 chars
	}
	return &Chunker{
		MaxChunkSize: maxSize,
		Overlap:      overlap,
	}
}

// SplitSection breaks a large section into smaller chunks if necessary.
func (c *Chunker) SplitSection(section Section) []string {
	if len(section.Content) <= c.MaxChunkSize {
		return []string{section.Content}
	}

	var chunks []string
	content := section.Content
	
	// Simple splitting by characters, trying to respect paragraph boundaries
	for len(content) > 0 {
		if len(content) <= c.MaxChunkSize {
			chunks = append(chunks, content)
			break
		}

		// Look for the last newline within the limit to avoid cutting mid-paragraph
		splitIdx := strings.LastIndex(content[:c.MaxChunkSize], "\n")
		if splitIdx == -1 || splitIdx < c.MaxChunkSize/2 {
			// If no newline found or it's too early, split at MaxChunkSize
			splitIdx = c.MaxChunkSize
		}

		chunks = append(chunks, content[:splitIdx])
		
		// Move forward with overlap
		nextStart := splitIdx - c.Overlap
		if nextStart < 0 {
			nextStart = splitIdx
		}
		
		content = content[nextStart:]
		if len(content) <= c.Overlap {
			break
		}
	}

	return chunks
}

// CreateChunks takes a list of sections and turns them into a list of model.Chunks.
func (c *Chunker) CreateChunks(repoID, filePath string, sections []Section) []models.Chunk {
	var allChunks []models.Chunk

	for _, s := range sections {
		splitTexts := c.SplitSection(s)
		
		for i, text := range splitTexts {
			// Generate a unique ID for the chunk
			// In a real app, we'd use a more robust hashing strategy
			chunk := models.Chunk{
				RepoID:      repoID,
				FilePath:    filePath,
				Content:     text,
				HeadingPath: strings.Join(s.HeadingPath, " > "),
				ChunkType:   models.ChunkTypeText,
				StartLine:   s.StartLine, // Note: line numbers are approximate for split chunks
				EndLine:     s.EndLine,
				Metadata: map[string]string{
					"part": fmt.Sprintf("%d/%d", i+1, len(splitTexts)),
				},
			}
			
			// Detect if content looks like code
			if strings.Contains(text, "```") {
				chunk.ChunkType = models.ChunkTypeCode
			}

			allChunks = append(allChunks, chunk)
		}
	}

	return allChunks
}

// fmt.Sprintf used above requires "fmt" import
