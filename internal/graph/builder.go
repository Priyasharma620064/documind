package graph

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/priya-sharma/documind/internal/models"
	"github.com/priya-sharma/documind/internal/storage"
)

// Builder populates the knowledge graph from stored data.
type Builder struct {
	db *storage.DB
}

// NewBuilder creates a new Builder.
func NewBuilder(db *storage.DB) *Builder {
	return &Builder{db: db}
}

// Build constructs the graph by scanning all repositories and files.
func (b *Builder) Build(ctx context.Context, g *Graph) error {
	slog.Info("Building knowledge graph...")

	// 1. Get all repositories
	repos, err := b.db.ListRepositories(ctx)
	if err != nil {
		return fmt.Errorf("listing repositories: %w", err)
	}

	for _, repo := range repos {
		repoNode := &models.GraphNode{
			ID:     repo.ID,
			Type:   models.NodeTypeFeature, // Using Feature as a proxy for the repo root for now
			Name:   repo.Name,
			RepoID: repo.ID,
			Metadata: map[string]string{
				"url":    repo.URL,
				"branch": repo.Branch,
			},
		}
		g.AddNode(repoNode)

		// 2. Get all files for this repo
		files, err := b.db.GetFilesByRepo(ctx, repo.ID)
		if err != nil {
			slog.Warn("Failed to get files for repo", "repo", repo.Name, "error", err)
			continue
		}

		for _, f := range files {
			fileNode := &models.GraphNode{
				ID:     f.ID,
				Type:   models.NodeTypeDocument,
				Name:   f.Path,
				RepoID: repo.ID,
				Metadata: map[string]string{
					"extension": f.Extension,
				},
			}
			g.AddNode(fileNode)

			// 3. Link file to repo
			g.AddEdge(repo.ID, f.ID, models.EdgeTypeDocuments, 1.0)
		}
	}

	slog.Info("Knowledge graph built", "nodes", len(g.Nodes), "edges", len(g.Edges))
	return nil
}
