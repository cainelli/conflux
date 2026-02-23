package planner

import (
	"fmt"

	"github.com/cainelli/conflux/internal/confluence"
	"github.com/cainelli/conflux/pkg/models"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Planner detects changes between local and remote pages
type Planner struct {
	client *confluence.Client
}

// NewPlanner creates a new planner
func NewPlanner(client *confluence.Client) *Planner {
	return &Planner{
		client: client,
	}
}

// Plan generates a list of changes between local and remote state
func (p *Planner) Plan(pages []*models.Page) ([]*models.Change, error) {
	var changes []*models.Change

	for _, page := range pages {
		change, err := p.detectChange(page)
		if err != nil {
			return nil, fmt.Errorf("failed to detect change for %s: %w", page.FilePath, err)
		}

		if change.Type != models.ChangeTypeNone {
			changes = append(changes, change)
		}
	}

	return changes, nil
}

// detectChange detects what kind of change is needed for a page
func (p *Planner) detectChange(page *models.Page) (*models.Change, error) {
	change := &models.Change{
		Page: page,
	}

	// Convert local content to Confluence Storage Format
	localContent, err := p.convertToStorage(page.Content, page.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to convert content: %w", err)
	}
	change.LocalContent = localContent

	// Check if page exists remotely
	if page.Confluence.PageID == "" {
		// New page
		change.Type = models.ChangeTypeCreate
		return change, nil
	}

	// Fetch remote page
	remotePage, err := p.client.GetPage(page.Confluence.PageID)
	if err != nil {
		// If page not found, treat as create
		change.Type = models.ChangeTypeCreate
		page.Confluence.PageID = "" // Clear invalid ID
		return change, nil
	}

	page.RemotePage = remotePage
	change.RemoteContent = remotePage.Content

	// Compare content
	if localContent != remotePage.Content {
		change.Type = models.ChangeTypeUpdate
		change.Diff = generateDiff(remotePage.Content, localContent)
		return change, nil
	}

	// Compare title
	if page.Title != remotePage.Title {
		change.Type = models.ChangeTypeUpdate
		return change, nil
	}

	// Compare labels
	if !equalStringSlices(page.Labels, remotePage.Labels) {
		change.Type = models.ChangeTypeUpdate
		return change, nil
	}

	// No changes
	change.Type = models.ChangeTypeNone
	return change, nil
}

// convertToStorage converts content to Confluence Storage Format.
func (p *Planner) convertToStorage(content string, _ models.ContentType) (string, error) {
	return content, nil
}

// generateDiff generates a human-readable diff
func generateDiff(old, new string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(old, new, false)
	return dmp.DiffPrettyText(diffs)
}

// equalStringSlices checks if two string slices contain the same elements
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}

	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}

	return true
}
