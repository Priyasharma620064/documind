package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Cloner handles Git repository cloning and updating.
type Cloner struct {
	baseDir string
	depth   int
}

// NewCloner creates a new Cloner that stores repos under baseDir.
func NewCloner(baseDir string, depth int) *Cloner {
	return &Cloner{
		baseDir: baseDir,
		depth:   depth,
	}
}

// CloneResult contains information about a clone/pull operation.
type CloneResult struct {
	LocalPath  string
	CommitHash string
	IsUpdated  bool
}

// CloneOrPull clones a repository if it doesn't exist locally,
// or pulls the latest changes if it does. Returns the local path
// and current HEAD commit hash.
func (c *Cloner) CloneOrPull(ctx context.Context, repoURL, branch string) (*CloneResult, error) {
	repoName := extractRepoName(repoURL)
	localPath := filepath.Join(c.baseDir, repoName)

	// Check if repo already exists locally
	if _, err := os.Stat(filepath.Join(localPath, ".git")); err == nil {
		return c.pull(ctx, localPath, branch)
	}

	return c.clone(ctx, repoURL, branch, localPath)
}

// clone performs a fresh clone of the repository.
func (c *Cloner) clone(ctx context.Context, repoURL, branch, localPath string) (*CloneResult, error) {
	slog.Info("Cloning repository",
		"url", repoURL,
		"branch", branch,
		"path", localPath,
	)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating parent directory: %w", err)
	}

	opts := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	// Set branch if specified
	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		opts.SingleBranch = true
	}

	// Use shallow clone if depth is configured
	if c.depth > 0 {
		opts.Depth = c.depth
	}

	repo, err := git.PlainCloneContext(ctx, localPath, false, opts)
	if err != nil {
		return nil, fmt.Errorf("cloning %s: %w", repoURL, err)
	}

	hash, err := getHeadHash(repo)
	if err != nil {
		return nil, err
	}

	slog.Info("Clone complete",
		"url", repoURL,
		"commit", hash,
	)

	return &CloneResult{
		LocalPath:  localPath,
		CommitHash: hash,
		IsUpdated:  true,
	}, nil
}

// pull updates an existing local repository.
func (c *Cloner) pull(ctx context.Context, localPath, branch string) (*CloneResult, error) {
	slog.Info("Pulling updates",
		"path", localPath,
		"branch", branch,
	)

	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, fmt.Errorf("opening repo at %s: %w", localPath, err)
	}

	// Get current hash before pull
	oldHash, err := getHeadHash(repo)
	if err != nil {
		return nil, err
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting worktree: %w", err)
	}

	pullOpts := &git.PullOptions{
		Progress: os.Stdout,
	}

	if branch != "" {
		pullOpts.ReferenceName = plumbing.NewBranchReferenceName(branch)
	}

	err = w.PullContext(ctx, pullOpts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("pulling %s: %w", localPath, err)
	}

	newHash, err := getHeadHash(repo)
	if err != nil {
		return nil, err
	}

	isUpdated := oldHash != newHash
	if isUpdated {
		slog.Info("Repository updated",
			"path", localPath,
			"old_commit", oldHash,
			"new_commit", newHash,
		)
	} else {
		slog.Info("Repository already up to date",
			"path", localPath,
			"commit", newHash,
		)
	}

	return &CloneResult{
		LocalPath:  localPath,
		CommitHash: newHash,
		IsUpdated:  isUpdated,
	}, nil
}

// getHeadHash returns the current HEAD commit hash.
func getHeadHash(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("getting HEAD: %w", err)
	}
	return head.Hash().String()[:12], nil
}

// extractRepoName extracts a directory name from a Git URL.
// e.g., "https://github.com/openkruise/kruise" → "openkruise-kruise"
func extractRepoName(url string) string {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")

	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "-" + parts[len(parts)-1]
	}
	if len(parts) >= 1 {
		return parts[len(parts)-1]
	}
	return "unknown"
}
