package search

import (
	"context"
	"fmt"
	"sort"

	"github.com/priya-sharma/documind/internal/embedding"
	"github.com/priya-sharma/documind/internal/models"
	"github.com/priya-sharma/documind/internal/vectorstore"
)

// Engine handles semantic and hybrid search across repositories.
type Engine struct {
	vectorStore vectorstore.VectorStore
	embedder    *embedding.OllamaClient
}

// NewEngine creates a new search Engine.
func NewEngine(vs vectorstore.VectorStore, embedder *embedding.OllamaClient) *Engine {
	return &Engine{
		vectorStore: vs,
		embedder:    embedder,
	}
}

// Search performs a semantic search with optional version filtering.
func (e *Engine) Search(ctx context.Context, repoName, query string, topK int, versionConstraint string) ([]models.SearchResult, error) {
	// 1. Search the vector store
	// In this implementation, we use the repoName as the collection name
	results, err := e.vectorStore.Search(ctx, repoName, query, topK*2) // Get more results to account for filtering
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 2. Apply version filtering if requested
	if versionConstraint != "" {
		filter, err := NewVersionFilter(versionConstraint)
		if err != nil {
			return nil, err
		}

		var filtered []models.SearchResult
		for _, res := range results {
			// In a real app, version would be in metadata
			version, ok := res.Chunk.Metadata["version"]
			if !ok || filter.Match(version) {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	// 3. Sort and limit
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}
