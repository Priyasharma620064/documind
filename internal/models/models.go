package models

import "time"

// Repository represents a tracked Git repository.
type Repository struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	URL          string    `json:"url" db:"url"`
	Branch       string    `json:"branch" db:"branch"`
	LocalPath    string    `json:"local_path" db:"local_path"`
	LastIngested time.Time `json:"last_ingested" db:"last_ingested"`
	CommitHash   string    `json:"commit_hash" db:"commit_hash"`
	Status       string    `json:"status" db:"status"` // pending, indexing, ready, error
}

// FileInfo represents a tracked file within a repository.
type FileInfo struct {
	ID           string    `json:"id" db:"id"`
	RepoID       string    `json:"repo_id" db:"repo_id"`
	Path         string    `json:"path" db:"path"`
	ContentHash  string    `json:"content_hash" db:"content_hash"`
	Size         int64     `json:"size" db:"size"`
	Extension    string    `json:"extension" db:"extension"`
	LastModified time.Time `json:"last_modified" db:"last_modified"`
	LastIndexed  time.Time `json:"last_indexed" db:"last_indexed"`
}

// Document represents a parsed document from a file.
type Document struct {
	ID       string            `json:"id"`
	RepoID   string            `json:"repo_id"`
	FilePath string            `json:"file_path"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	Chunks   []Chunk           `json:"chunks"`
	Metadata map[string]string `json:"metadata"`
}

// Chunk represents a semantic unit of text for embedding.
type Chunk struct {
	ID          string            `json:"id"`
	DocumentID  string            `json:"document_id"`
	RepoID      string            `json:"repo_id"`
	Content     string            `json:"content"`
	HeadingPath string            `json:"heading_path"` // e.g., "CloneSet > Scaling > Partition"
	ChunkType   ChunkType         `json:"chunk_type"`
	StartLine   int               `json:"start_line"`
	EndLine     int               `json:"end_line"`
	FilePath    string            `json:"file_path"`
	Metadata    map[string]string `json:"metadata"`
}

// ChunkType classifies the type of content in a chunk.
type ChunkType string

const (
	ChunkTypeText     ChunkType = "text"
	ChunkTypeCode     ChunkType = "code"
	ChunkTypeYAML     ChunkType = "yaml"
	ChunkTypeHeading  ChunkType = "heading"
	ChunkTypeFrontmatter ChunkType = "frontmatter"
)

// SearchResult represents a single search result.
type SearchResult struct {
	Chunk      Chunk   `json:"chunk"`
	Score      float64 `json:"score"`
	RepoName   string  `json:"repo_name"`
	Version    string  `json:"version,omitempty"`
}

// SearchQuery represents a search request.
type SearchQuery struct {
	Query      string            `json:"query"`
	TopK       int               `json:"top_k"`
	Filters    map[string]string `json:"filters,omitempty"`
	Version    string            `json:"version,omitempty"`
	RepoFilter string            `json:"repo_filter,omitempty"`
}

// QualityIssue represents a documentation quality problem.
type QualityIssue struct {
	ID          string        `json:"id"`
	Type        IssueType     `json:"type"`
	Severity    IssueSeverity `json:"severity"`
	FilePath    string        `json:"file_path"`
	Line        int           `json:"line,omitempty"`
	Description string        `json:"description"`
	Suggestion  string        `json:"suggestion,omitempty"`
	RepoID      string        `json:"repo_id"`
}

// IssueType classifies documentation quality issues.
type IssueType string

const (
	IssueTypeBrokenLink     IssueType = "broken_link"
	IssueTypeStaleContent   IssueType = "stale_content"
	IssueTypeInvalidYAML    IssueType = "invalid_yaml"
	IssueTypeDuplicate      IssueType = "duplicate_content"
	IssueTypeDeprecatedAPI  IssueType = "deprecated_api"
	IssueTypeMissingDoc     IssueType = "missing_documentation"
)

// IssueSeverity indicates how critical an issue is.
type IssueSeverity string

const (
	SeverityCritical IssueSeverity = "critical"
	SeverityWarning  IssueSeverity = "warning"
	SeverityInfo     IssueSeverity = "info"
)

// QualityReport is the output of a documentation evaluation run.
type QualityReport struct {
	RepoID      string         `json:"repo_id"`
	RepoName    string         `json:"repo_name"`
	GeneratedAt time.Time      `json:"generated_at"`
	TotalFiles  int            `json:"total_files"`
	Issues      []QualityIssue `json:"issues"`
	Summary     ReportSummary  `json:"summary"`
}

// ReportSummary provides aggregate quality metrics.
type ReportSummary struct {
	TotalIssues    int     `json:"total_issues"`
	CriticalCount  int     `json:"critical_count"`
	WarningCount   int     `json:"warning_count"`
	InfoCount      int     `json:"info_count"`
	HealthScore    float64 `json:"health_score"` // 0.0 - 1.0
}

// ChangeSet represents differences between two ingestion runs.
type ChangeSet struct {
	Added    []FileInfo `json:"added"`
	Modified []FileInfo `json:"modified"`
	Deleted  []FileInfo `json:"deleted"`
}

// GraphNode represents a node in the knowledge graph.
type GraphNode struct {
	ID       string            `json:"id"`
	Type     NodeType          `json:"type"`
	Name     string            `json:"name"`
	RepoID   string            `json:"repo_id"`
	Metadata map[string]string `json:"metadata"`
}

// NodeType classifies knowledge graph nodes.
type NodeType string

const (
	NodeTypeFeature  NodeType = "feature"
	NodeTypeDocument NodeType = "document"
	NodeTypeCodeFile NodeType = "code_file"
	NodeTypeRelease  NodeType = "release"
	NodeTypeExample  NodeType = "example"
	NodeTypeCRD      NodeType = "crd"
)

// GraphEdge represents a relationship between two nodes.
type GraphEdge struct {
	SourceID string   `json:"source_id"`
	TargetID string   `json:"target_id"`
	Type     EdgeType `json:"type"`
	Weight   float64  `json:"weight"`
}

// EdgeType classifies knowledge graph edges.
type EdgeType string

const (
	EdgeTypeDocuments  EdgeType = "documents"
	EdgeTypeImplements EdgeType = "implements"
	EdgeTypeExamples   EdgeType = "examples"
	EdgeTypeReleasedIn EdgeType = "released_in"
	EdgeTypeDependsOn  EdgeType = "depends_on"
	EdgeTypeRelatedTo  EdgeType = "related_to"
)
