package vectorstore

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/philippgille/chromem-go"
	"github.com/priya-sharma/documind/internal/models"
)

// VectorStore defines the interface for storing and searching embeddings.
type VectorStore interface {
	AddDocuments(ctx context.Context, collectionName string, docs []models.Chunk) error
	Search(ctx context.Context, collectionName string, query string, topK int) ([]models.SearchResult, error)
}

// ChromemStore implements VectorStore using chromem-go.
type ChromemStore struct {
	db *chromem.DB
}

// NewChromemStore creates a new ChromemStore.
func NewChromemStore() *ChromemStore {
	return &ChromemStore{
		db: chromem.NewDB(),
	}
}

// AddDocuments adds a slice of chunks to a specific collection.
func (s *ChromemStore) AddDocuments(ctx context.Context, collectionName string, chunks []models.Chunk) error {
	col, err := s.db.CreateCollection(collectionName, nil, nil)
	if err != nil {
		// If it already exists, just get it
		col = s.db.GetCollection(collectionName, nil)
		if col == nil {
			return fmt.Errorf("failed to create or get collection %s", collectionName)
		}
	}

	var docs []chromem.Document
	for _, chunk := range chunks {
		docs = append(docs, chromem.Document{
			ID:      chunk.ID,
			Content: chunk.Content,
			Metadata: map[string]string{
				"repo_id":      chunk.RepoID,
				"file_path":    chunk.FilePath,
				"heading_path": chunk.HeadingPath,
				"type":         string(chunk.ChunkType),
			},
		})
	}

	if err := col.AddDocuments(ctx, docs, 0); err != nil {
		return fmt.Errorf("adding documents to chromem: %w", err)
	}

	slog.Info("Added documents to vector store", "collection", collectionName, "count", len(docs))
	return nil
}

// Search performs semantic search on a collection.
func (s *ChromemStore) Search(ctx context.Context, collectionName string, query string, topK int) ([]models.SearchResult, error) {
	col := s.db.GetCollection(collectionName, nil)
	if col == nil {
		return nil, fmt.Errorf("collection %s not found", collectionName)
	}

	results, err := col.Query(ctx, query, topK, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("querying chromem: %w", err)
	}

	var searchResults []models.SearchResult
	for _, res := range results {
		searchResults = append(searchResults, models.SearchResult{
			Chunk: models.Chunk{
				ID:          res.ID,
				Content:     res.Content,
				RepoID:      res.Metadata["repo_id"],
				FilePath:    res.Metadata["file_path"],
				HeadingPath: res.Metadata["heading_path"],
				ChunkType:   models.ChunkType(res.Metadata["type"]),
			},
			Score: float64(res.Similarity),
		})
	}

	return searchResults, nil
}
