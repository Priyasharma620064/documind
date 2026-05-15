package search

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// VersionFilter handles semantic version matching.
type VersionFilter struct {
	constraint *semver.Constraints
}

// NewVersionFilter creates a new VersionFilter from a semver constraint string.
// e.g., ">=1.5, <1.8"
func NewVersionFilter(constraintStr string) (*VersionFilter, error) {
	c, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return nil, fmt.Errorf("invalid semver constraint: %w", err)
	}
	return &VersionFilter{constraint: c}, nil
}

// Match checks if a specific version matches the filter.
func (f *VersionFilter) Match(versionStr string) bool {
	v, err := semver.NewVersion(versionStr)
	if err != nil {
		return false
	}
	return f.constraint.Check(v)
}
