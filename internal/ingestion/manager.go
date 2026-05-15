package ingestion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"github.com/priya-sharma/documind/internal/config"
	"github.com/priya-sharma/documind/internal/storage"
)

// Manager orchestrates the full ingestion pipeline:
// clone/pull → walk → diff → store state.
type Manager struct {
	cloner  *Cloner
	walker  *Walker
	differ  *Differ
	db      *storage.DB
	cfg     *config.Config
}

// NewManager creates a new ingestion Manager.
func NewManager(db *storage.DB, cfg *config.Config) *Manager {
	return &Manager{
		cloner: NewCloner(
			cfg.Storage.DataDir+"/repos",
			cfg.Ingestion.CloneDepth,
		),
		walker: NewWalker(
			cfg.Ingestion.FileExtensions,
			cfg.Ingestion.MaxFileSizeMB,
		),
		differ: NewDiffer(),
		db:     db,
		cfg:    cfg,
	}
}

// IngestResult contains the results of an ingestion run.
type IngestResult struct {
	RepoName      string
	CommitHash    string
	FilesAdded    int
	FilesModified int
	FilesDeleted  int
	TotalFiles    int
	Duration      time.Duration
	IsUpdated     bool
}

// IngestRepo ingests a single repository by URL and branch.
func (m *Manager) IngestRepo(ctx context.Context, name, repoURL, branch string) (*IngestResult, error) {
	startTime := time.Now()

	slog.Info("Starting ingestion",
		"repo", name,
		"url", repoURL,
		"branch", branch,
	)

	// Generate a deterministic repo ID
	repoID := generateID(repoURL + "/" + branch)

	// Log the ingestion run
	ingestionLog := &storage.IngestionLog{
		RepoID:    repoID,
		StartedAt: startTime,
		Status:    "running",
	}
	logID, err := m.db.InsertIngestionLog(ctx, ingestionLog)
	if err != nil {
		slog.Warn("Failed to create ingestion log", "error", err)
	}

	// Step 1: Clone or pull
	cloneResult, err := m.cloner.CloneOrPull(ctx, repoURL, branch)
	if err != nil {
		m.failIngestionLog(ctx, logID, err)
		return nil, fmt.Errorf("clone/pull failed: %w", err)
	}

	// Step 2: Update repo state in DB
	repo := &storage.Repository{
		ID:           repoID,
		Name:         name,
		URL:          repoURL,
		Branch:       branch,
		LocalPath:    cloneResult.LocalPath,
		CommitHash:   cloneResult.CommitHash,
		LastIngested: time.Now(),
		Status:       "indexing",
	}
	if err := m.db.UpsertRepository(ctx, repo); err != nil {
		return nil, fmt.Errorf("updating repo state: %w", err)
	}

	// Step 3: Walk the repository
	currentFiles, err := m.walker.Walk(cloneResult.LocalPath)
	if err != nil {
		m.failIngestionLog(ctx, logID, err)
		return nil, fmt.Errorf("walking repo: %w", err)
	}

	// Step 4: Get previous file state from DB
	previousDBFiles, err := m.db.GetFilesByRepo(ctx, repoID)
	if err != nil {
		slog.Warn("No previous state found, treating as fresh ingestion", "error", err)
	}

	// Convert DB files to WalkedFile format for diffing
	previousFiles := dbFilesToWalked(previousDBFiles)

	// Step 5: Compute diff
	changes := m.differ.Diff(previousFiles, currentFiles)

	// Step 6: Update file state in DB
	for _, f := range currentFiles {
		fileID := generateID(repoID + "/" + f.Path)
		dbFile := &storage.File{
			ID:           fileID,
			RepoID:       repoID,
			Path:         f.Path,
			ContentHash:  f.ContentHash,
			Size:         f.Size,
			Extension:    f.Extension,
			LastModified: f.ModTime,
			LastIndexed:  time.Now(),
		}
		if err := m.db.UpsertFile(ctx, dbFile); err != nil {
			slog.Warn("Failed to upsert file", "path", f.Path, "error", err)
		}
	}

	// Update repo status
	repo.Status = "ready"
	if err := m.db.UpsertRepository(ctx, repo); err != nil {
		slog.Warn("Failed to update repo status", "error", err)
	}

	duration := time.Since(startTime)
	result := &IngestResult{
		RepoName:      name,
		CommitHash:    cloneResult.CommitHash,
		FilesAdded:    len(changes.Added),
		FilesModified: len(changes.Modified),
		FilesDeleted:  len(changes.Deleted),
		TotalFiles:    len(currentFiles),
		Duration:      duration,
		IsUpdated:     changes.HasChanges(),
	}

	// Complete the ingestion log
	if logID > 0 {
		completedAt := time.Now()
		ingestionLog.CompletedAt = &completedAt
		ingestionLog.Status = "completed"
		ingestionLog.FilesAdded = result.FilesAdded
		ingestionLog.FilesModified = result.FilesModified
		ingestionLog.FilesDeleted = result.FilesDeleted
		_ = m.db.UpdateIngestionLog(ctx, logID, ingestionLog)
	}

	slog.Info("Ingestion complete",
		"repo", name,
		"commit", cloneResult.CommitHash,
		"total_files", result.TotalFiles,
		"added", result.FilesAdded,
		"modified", result.FilesModified,
		"deleted", result.FilesDeleted,
		"duration", duration.Round(time.Millisecond),
	)

	return result, nil
}

// IngestAll ingests all repositories defined in config.
func (m *Manager) IngestAll(ctx context.Context) ([]*IngestResult, error) {
	var results []*IngestResult

	for _, repoCfg := range m.cfg.Repositories {
		for _, branch := range repoCfg.Branches {
			result, err := m.IngestRepo(ctx, repoCfg.Name, repoCfg.URL, branch)
			if err != nil {
				slog.Error("Failed to ingest repo",
					"repo", repoCfg.Name,
					"branch", branch,
					"error", err,
				)
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// GetChangedFiles returns the files that need re-indexing for a repository.
// This is used by downstream pipeline stages (parsing, embedding).
func (m *Manager) GetChangedFiles(ctx context.Context, repoName string) (*ChangeSet, error) {
	repo, err := m.db.GetRepository(ctx, repoName)
	if err != nil {
		return nil, fmt.Errorf("repository not found: %s", repoName)
	}

	currentFiles, err := m.walker.Walk(repo.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("walking repo: %w", err)
	}

	previousDBFiles, err := m.db.GetFilesByRepo(ctx, repo.ID)
	if err != nil {
		return nil, fmt.Errorf("getting previous state: %w", err)
	}

	previousFiles := dbFilesToWalked(previousDBFiles)
	return m.differ.Diff(previousFiles, currentFiles), nil
}

// failIngestionLog marks an ingestion log as failed.
func (m *Manager) failIngestionLog(ctx context.Context, logID int64, ingestionErr error) {
	if logID <= 0 {
		return
	}
	completedAt := time.Now()
	errMsg := ingestionErr.Error()
	log := &storage.IngestionLog{
		CompletedAt:  &completedAt,
		Status:       "failed",
		ErrorMessage: &errMsg,
	}
	_ = m.db.UpdateIngestionLog(ctx, logID, log)
}

// dbFilesToWalked converts storage file records to WalkedFile format.
func dbFilesToWalked(dbFiles []storage.File) []WalkedFile {
	files := make([]WalkedFile, len(dbFiles))
	for i, f := range dbFiles {
		files[i] = WalkedFile{
			Path:        f.Path,
			ContentHash: f.ContentHash,
			Size:        f.Size,
			Extension:   f.Extension,
			ModTime:     f.LastModified,
		}
	}
	return files
}

// generateID creates a deterministic short ID from a string.
func generateID(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])[:16]
}
