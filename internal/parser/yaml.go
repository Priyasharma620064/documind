package parser

import (
	"fmt"
	"strings"

	"github.com/priya-sharma/documind/internal/models"
	"gopkg.in/yaml.v3"
)

// YAMLParser parses YAML files, specifically targeting Kubernetes resources.
type YAMLParser struct{}

// NewYAMLParser creates a new YAMLParser.
func NewYAMLParser() *YAMLParser {
	return &YAMLParser{}
}

// Parse extracts key information from a YAML document.
func (p *YAMLParser) Parse(repoID, filePath string, source []byte) ([]models.Chunk, error) {
	// Support multi-document YAML files
	decoder := yaml.NewDecoder(strings.NewReader(string(source)))
	var chunks []models.Chunk

	for {
		var raw map[string]interface{}
		if err := decoder.Decode(&raw); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("decoding YAML: %w", err)
		}

		// Extract K8s specific fields if they exist
		kind, _ := raw["kind"].(string)
		apiVersion, _ := raw["apiVersion"].(string)
		metadata, _ := raw["metadata"].(map[string]interface{})
		name, _ := metadata["name"].(string)

		if kind == "" {
			kind = "YAML"
		}

		title := kind
		if name != "" {
			title = fmt.Sprintf("%s: %s", kind, name)
		}

		// Re-encode to keep the specific document content
		content, _ := yaml.Marshal(raw)

		chunk := models.Chunk{
			RepoID:      repoID,
			FilePath:    filePath,
			Content:     string(content),
			HeadingPath: title,
			ChunkType:   models.ChunkTypeYAML,
			Metadata: map[string]string{
				"kind":       kind,
				"apiVersion": apiVersion,
				"name":       name,
			},
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}
