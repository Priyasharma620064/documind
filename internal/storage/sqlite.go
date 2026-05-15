package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite database connection for metadata storage.
type DB struct {
	conn *sql.DB
	path string
}

// New creates a new SQLite database connection and initializes the schema.
func New(dbPath string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(1) // SQLite only supports one writer
	conn.SetMaxIdleConns(1)
	conn.SetConnMaxLifetime(0)

	db := &DB{conn: conn, path: dbPath}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("Database initialized", "path", dbPath)
	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate creates or updates the database schema.
func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS repositories (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL,
			branch TEXT NOT NULL DEFAULT 'main',
			local_path TEXT,
			last_ingested DATETIME,
			commit_hash TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL,
			path TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			extension TEXT,
			last_modified DATETIME,
			last_indexed DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE,
			UNIQUE(repo_id, path)
		)`,

		`CREATE TABLE IF NOT EXISTS ingestion_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repo_id TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			completed_at DATETIME,
			status TEXT NOT NULL DEFAULT 'running',
			files_added INTEGER DEFAULT 0,
			files_modified INTEGER DEFAULT 0,
			files_deleted INTEGER DEFAULT 0,
			chunks_created INTEGER DEFAULT 0,
			error_message TEXT,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
		)`,

		`CREATE TABLE IF NOT EXISTS quality_issues (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL,
			type TEXT NOT NULL,
			severity TEXT NOT NULL,
			file_path TEXT NOT NULL,
			line INTEGER,
			description TEXT NOT NULL,
			suggestion TEXT,
			detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME,
			FOREIGN KEY (repo_id) REFERENCES repositories(id) ON DELETE CASCADE
		)`,

		// Indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_files_repo_id ON files(repo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_files_content_hash ON files(content_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_ingestion_logs_repo_id ON ingestion_logs(repo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_quality_issues_repo_id ON quality_issues(repo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_quality_issues_type ON quality_issues(type)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("executing migration: %w\nSQL: %s", err, m)
		}
	}

	return nil
}

// UpsertRepository inserts or updates a repository record.
func (db *DB) UpsertRepository(ctx context.Context, repo *Repository) error {
	query := `
		INSERT INTO repositories (id, name, url, branch, local_path, last_ingested, commit_hash, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			url = excluded.url,
			branch = excluded.branch,
			local_path = excluded.local_path,
			last_ingested = excluded.last_ingested,
			commit_hash = excluded.commit_hash,
			status = excluded.status,
			updated_at = CURRENT_TIMESTAMP`

	_, err := db.conn.ExecContext(ctx, query,
		repo.ID, repo.Name, repo.URL, repo.Branch,
		repo.LocalPath, repo.LastIngested, repo.CommitHash, repo.Status,
	)
	return err
}

// GetRepository returns a repository by name.
func (db *DB) GetRepository(ctx context.Context, name string) (*Repository, error) {
	query := `SELECT id, name, url, branch, local_path, last_ingested, commit_hash, status
		FROM repositories WHERE name = ?`

	var repo Repository
	var lastIngested sql.NullTime
	var localPath, commitHash sql.NullString

	err := db.conn.QueryRowContext(ctx, query, name).Scan(
		&repo.ID, &repo.Name, &repo.URL, &repo.Branch,
		&localPath, &lastIngested, &commitHash, &repo.Status,
	)
	if err != nil {
		return nil, err
	}

	if localPath.Valid {
		repo.LocalPath = localPath.String
	}
	if lastIngested.Valid {
		repo.LastIngested = lastIngested.Time
	}
	if commitHash.Valid {
		repo.CommitHash = commitHash.String
	}

	return &repo, nil
}

// ListRepositories returns all tracked repositories.
func (db *DB) ListRepositories(ctx context.Context) ([]Repository, error) {
	query := `SELECT id, name, url, branch, local_path, last_ingested, commit_hash, status
		FROM repositories ORDER BY name`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		var lastIngested sql.NullTime
		var localPath, commitHash sql.NullString

		if err := rows.Scan(
			&repo.ID, &repo.Name, &repo.URL, &repo.Branch,
			&localPath, &lastIngested, &commitHash, &repo.Status,
		); err != nil {
			return nil, err
		}

		if localPath.Valid {
			repo.LocalPath = localPath.String
		}
		if lastIngested.Valid {
			repo.LastIngested = lastIngested.Time
		}
		if commitHash.Valid {
			repo.CommitHash = commitHash.String
		}
		repos = append(repos, repo)
	}

	return repos, rows.Err()
}

// UpsertFile inserts or updates a file record.
func (db *DB) UpsertFile(ctx context.Context, f *File) error {
	query := `
		INSERT INTO files (id, repo_id, path, content_hash, size, extension, last_modified, last_indexed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_id, path) DO UPDATE SET
			content_hash = excluded.content_hash,
			size = excluded.size,
			last_modified = excluded.last_modified,
			last_indexed = excluded.last_indexed`

	_, err := db.conn.ExecContext(ctx, query,
		f.ID, f.RepoID, f.Path, f.ContentHash,
		f.Size, f.Extension, f.LastModified, f.LastIndexed,
	)
	return err
}

// GetFilesByRepo returns all indexed files for a repository.
func (db *DB) GetFilesByRepo(ctx context.Context, repoID string) ([]File, error) {
	query := `SELECT id, repo_id, path, content_hash, size, extension, last_modified, last_indexed
		FROM files WHERE repo_id = ? ORDER BY path`

	rows, err := db.conn.QueryContext(ctx, query, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		if err := rows.Scan(
			&f.ID, &f.RepoID, &f.Path, &f.ContentHash,
			&f.Size, &f.Extension, &f.LastModified, &f.LastIndexed,
		); err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	return files, rows.Err()
}

// DeleteFilesByRepo removes all file records for a repository.
func (db *DB) DeleteFilesByRepo(ctx context.Context, repoID string) error {
	_, err := db.conn.ExecContext(ctx, `DELETE FROM files WHERE repo_id = ?`, repoID)
	return err
}

// InsertIngestionLog creates a new ingestion log entry.
func (db *DB) InsertIngestionLog(ctx context.Context, log *IngestionLog) (int64, error) {
	query := `INSERT INTO ingestion_logs (repo_id, started_at, status)
		VALUES (?, ?, ?)`

	result, err := db.conn.ExecContext(ctx, query, log.RepoID, log.StartedAt, log.Status)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateIngestionLog updates an existing ingestion log entry.
func (db *DB) UpdateIngestionLog(ctx context.Context, id int64, log *IngestionLog) error {
	query := `UPDATE ingestion_logs SET
		completed_at = ?, status = ?,
		files_added = ?, files_modified = ?, files_deleted = ?,
		chunks_created = ?, error_message = ?
		WHERE id = ?`

	_, err := db.conn.ExecContext(ctx, query,
		log.CompletedAt, log.Status,
		log.FilesAdded, log.FilesModified, log.FilesDeleted,
		log.ChunksCreated, log.ErrorMessage,
		id,
	)
	return err
}

// Repository is the storage model for a tracked repository.
type Repository struct {
	ID           string
	Name         string
	URL          string
	Branch       string
	LocalPath    string
	LastIngested time.Time
	CommitHash   string
	Status       string
}

// File is the storage model for a tracked file.
type File struct {
	ID           string
	RepoID       string
	Path         string
	ContentHash  string
	Size         int64
	Extension    string
	LastModified time.Time
	LastIndexed  time.Time
}

// IngestionLog is the storage model for an ingestion run.
type IngestionLog struct {
	ID            int64
	RepoID        string
	StartedAt     time.Time
	CompletedAt   *time.Time
	Status        string
	FilesAdded    int
	FilesModified int
	FilesDeleted  int
	ChunksCreated int
	ErrorMessage  *string
}
