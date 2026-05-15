package version

// Build-time variables injected via ldflags.
var (
	// Version is the semantic version tag (e.g., v0.1.0).
	Version = "dev"

	// Commit is the short Git commit hash.
	Commit = "none"

	// BuildTime is the UTC timestamp of the build.
	BuildTime = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return Version + " (commit: " + Commit + ", built: " + BuildTime + ")"
}
