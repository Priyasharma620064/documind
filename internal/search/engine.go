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

// Search performs a semantic search.
func (e *Engine) Search(ctx context.Context, repoName, query string, topK int) ([]models.SearchResult, error) {
	// 1. Search the vector store
	// In this implementation, we use the repoName as the collection name
	results, err := e.vectorStore.Search(ctx, repoName, query, topK)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 3. Sort by score descending (chromem-go already does this, but being safe)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}
