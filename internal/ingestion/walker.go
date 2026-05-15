package ingestion

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WalkedFile represents a file discovered during repository walking.
type WalkedFile struct {
	Path         string    // Relative path from repo root
	AbsolutePath string    // Full filesystem path
	Size         int64     // File size in bytes
	Extension    string    // File extension (e.g., ".md")
	ContentHash  string    // SHA-256 hash of file content
	ModTime      time.Time // Last modification time
}

// Walker traverses a repository directory and discovers indexable files.
type Walker struct {
	extensions   map[string]bool
	maxSizeBytes int64
}

// NewWalker creates a Walker that filters by the given file extensions
// and maximum file size in megabytes.
func NewWalker(extensions []string, maxSizeMB int) *Walker {
	extMap := make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extMap[ext] = true
	}

	return &Walker{
		extensions:   extMap,
		maxSizeBytes: int64(maxSizeMB) * 1024 * 1024,
	}
}

// Walk traverses the repository at rootDir and returns all indexable files.
// It skips hidden directories, vendor directories, and files that don't
// match the configured extensions.
func (w *Walker) Walk(rootDir string) ([]WalkedFile, error) {
	var files []WalkedFile

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Warn("Error accessing path", "path", path, "error", err)
			return nil // Continue walking despite errors
		}

		// Skip hidden directories and common non-content directories
		if info.IsDir() {
			name := info.Name()
			if shouldSkipDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check file extension
		ext := strings.ToLower(filepath.Ext(path))
		if !w.extensions[ext] {
			// Also check for special filenames without extensions
			baseName := strings.ToUpper(info.Name())
			if baseName != "README" && baseName != "CHANGELOG" && baseName != "CONTRIBUTING" {
				return nil
			}
		}

		// Check file size
		if info.Size() > w.maxSizeBytes {
			slog.Debug("Skipping large file",
				"path", path,
				"size_mb", info.Size()/(1024*1024),
			)
			return nil
		}

		// Skip empty files
		if info.Size() == 0 {
			return nil
		}

		// Compute content hash
		hash, err := hashFile(path)
		if err != nil {
			slog.Warn("Error hashing file", "path", path, "error", err)
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			relPath = path
		}

		files = append(files, WalkedFile{
			Path:         relPath,
			AbsolutePath: path,
			Size:         info.Size(),
			Extension:    ext,
			ContentHash:  hash,
			ModTime:      info.ModTime(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory %s: %w", rootDir, err)
	}

	slog.Info("Walk complete",
		"root", rootDir,
		"files_found", len(files),
	)

	return files, nil
}

// shouldSkipDir returns true for directories that should not be traversed.
func shouldSkipDir(name string) bool {
	skipDirs := map[string]bool{
		".git":         true,
		".github":      false, // We may want workflow files later
		"node_modules": true,
		"vendor":       true,
		".idea":        true,
		".vscode":      true,
		"__pycache__":  true,
		"dist":         true,
		"build":        true,
		".cache":       true,
	}

	if strings.HasPrefix(name, ".") && name != ".github" {
		return true
	}

	return skipDirs[name]
}

// hashFile computes the SHA-256 hash of a file's content.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil))[:16], nil // First 16 hex chars
}
