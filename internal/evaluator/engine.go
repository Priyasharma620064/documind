package evaluator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/priya-sharma/documind/internal/models"
	"github.com/priya-sharma/documind/internal/storage"
)

// Engine runs documentation quality evaluations.
type Engine struct {
	db          *storage.DB
	linkChecker *LinkChecker
}

// NewEngine creates a new evaluation Engine.
func NewEngine(db *storage.DB) *Engine {
	return &Engine{
		db:          db,
		linkChecker: NewLinkChecker(),
	}
}

// EvaluateRepo runs all quality checks for a repository.
func (e *Engine) EvaluateRepo(ctx context.Context, repoID string) (*models.QualityReport, error) {
	// 1. Get all files for the repo
	files, err := e.db.GetFilesByRepo(ctx, repoID)
	if err != nil {
		return nil, fmt.Errorf("getting files: %w", err)
	}

	report := &models.QualityReport{
		RepoID:      repoID,
		GeneratedAt: time.Now(),
		TotalFiles:  len(files),
	}

	// 2. Iterate and check files
	for _, f := range files {
		// This is where we would read the file content and run checks.
		// For this MVP, we'll just log and move on, as we need a way 
		// to easily read file contents from the local path.
		slog.Debug("Evaluating file", "path", f.Path)
		
		// Example: Link checks would run here if we had the content
	}

	report.Summary = models.ReportSummary{
		TotalIssues: len(report.Issues),
	}

	return report, nil
}
