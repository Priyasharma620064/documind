package evaluator

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/priya-sharma/documind/internal/models"
)

// LinkChecker validates links in markdown content.
type LinkChecker struct {
	client *http.Client
}

// NewLinkChecker creates a new LinkChecker.
func NewLinkChecker() *LinkChecker {
	return &LinkChecker{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

var linkRegex = regexp.MustCompile(`\[.*?\]\((https?://.*?|/.*?|#.*?)\)`)

// Check finds and validates links in the given content.
func (c *LinkChecker) Check(repoID, filePath, content string) []models.QualityIssue {
	matches := linkRegex.FindAllStringSubmatch(content, -1)
	var issues []models.QualityIssue

	for _, match := range matches {
		link := match[1]
		if strings.HasPrefix(link, "http") {
			// External link check (async in a real app, but synchronous here for simplicity)
			if err := c.checkExternal(link); err != nil {
				issues = append(issues, models.QualityIssue{
					RepoID:      repoID,
					Type:        models.IssueTypeBrokenLink,
					Severity:    models.SeverityWarning,
					FilePath:    filePath,
					Description: fmt.Sprintf("Broken external link: %s (%v)", link, err),
					Suggestion:  "Check if the URL is still valid or has moved.",
				})
			}
		}
		// Internal link checking would go here
	}

	return issues
}

func (c *LinkChecker) checkExternal(url string) error {
	resp, err := c.client.Head(url)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return nil
}

// fmt.Sprintf used above requires "strings" import
