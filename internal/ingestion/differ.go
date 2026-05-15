package ingestion

import "log/slog"

// ChangeSet represents the differences between two ingestion runs.
type ChangeSet struct {
	Added    []WalkedFile
	Modified []WalkedFile
	Deleted  []WalkedFile
}

// Differ compares two sets of walked files to determine what changed.
type Differ struct{}

// NewDiffer creates a new Differ.
func NewDiffer() *Differ {
	return &Differ{}
}

// Diff computes the changes between a previous and current set of files.
// It uses content hashes to detect modifications — only files whose
// content actually changed are marked as modified.
func (d *Differ) Diff(previous, current []WalkedFile) *ChangeSet {
	prevMap := make(map[string]WalkedFile, len(previous))
	for _, f := range previous {
		prevMap[f.Path] = f
	}

	currMap := make(map[string]WalkedFile, len(current))
	for _, f := range current {
		currMap[f.Path] = f
	}

	cs := &ChangeSet{}

	// Find added and modified files
	for path, currFile := range currMap {
		prevFile, exists := prevMap[path]
		if !exists {
			cs.Added = append(cs.Added, currFile)
		} else if prevFile.ContentHash != currFile.ContentHash {
			cs.Modified = append(cs.Modified, currFile)
		}
	}

	// Find deleted files
	for path, prevFile := range prevMap {
		if _, exists := currMap[path]; !exists {
			cs.Deleted = append(cs.Deleted, prevFile)
		}
	}

	slog.Info("Diff complete",
		"added", len(cs.Added),
		"modified", len(cs.Modified),
		"deleted", len(cs.Deleted),
	)

	return cs
}

// HasChanges returns true if the changeset contains any changes.
func (cs *ChangeSet) HasChanges() bool {
	return len(cs.Added) > 0 || len(cs.Modified) > 0 || len(cs.Deleted) > 0
}

// TotalChanged returns the total number of files that changed.
func (cs *ChangeSet) TotalChanged() int {
	return len(cs.Added) + len(cs.Modified) + len(cs.Deleted)
}
