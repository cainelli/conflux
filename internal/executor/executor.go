package executor

import (
	"fmt"
	"sort"
	"time"

	"github.com/cainelli/conflux/internal/confluence"
	"github.com/cainelli/conflux/internal/local"
	"github.com/cainelli/conflux/pkg/models"
)

// Executor applies changes to Confluence
type Executor struct {
	client *confluence.Client
	writer *local.Writer
}

// NewExecutor creates a new executor
func NewExecutor(client *confluence.Client, writer *local.Writer) *Executor {
	return &Executor{
		client: client,
		writer: writer,
	}
}

// Execute applies changes to Confluence and updates local metadata
func (e *Executor) Execute(changes []*models.Change) error {
	// Sort changes by dependency (parents before children)
	sortedChanges, err := e.sortByDependency(changes)
	if err != nil {
		return fmt.Errorf("failed to sort changes: %w", err)
	}

	// Apply each change
	for _, change := range sortedChanges {
		if err := e.executeChange(change); err != nil {
			return fmt.Errorf("failed to execute change for %s: %w", change.Page.FilePath, err)
		}
	}

	return nil
}

// executeChange applies a single change
func (e *Executor) executeChange(change *models.Change) error {
	page := change.Page

	switch change.Type {
	case models.ChangeTypeCreate:
		return e.createPage(page, change.LocalContent)

	case models.ChangeTypeUpdate:
		return e.updatePage(page, change.LocalContent)

	case models.ChangeTypeDelete:
		return fmt.Errorf("delete operation not implemented")

	default:
		return nil
	}
}

// createPage creates a new page in Confluence
func (e *Executor) createPage(page *models.Page, content string) error {
	// Resolve parent ID
	parentID := ""
	if page.Parent != "" {
		// If parent is a page ID, use it directly
		if isPageID(page.Parent) {
			parentID = page.Parent
		} else {
			return fmt.Errorf("parent must be a page ID for create operations: %s", page.Parent)
		}
	}

	// Create page
	remotePage, err := e.client.CreatePage(
		page.Confluence.SpaceKey,
		page.Title,
		content,
		parentID,
		page.Labels,
	)
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}

	// Update local metadata
	page.Confluence.PageID = remotePage.ID
	page.Version = remotePage.Version
	page.LastSync = time.Now()

	// Write updated metadata back to file
	if err := e.writer.UpdatePageMetadata(page); err != nil {
		return fmt.Errorf("failed to update local metadata: %w", err)
	}

	return nil
}

// updatePage updates an existing page in Confluence
func (e *Executor) updatePage(page *models.Page, content string) error {
	if page.Confluence.PageID == "" {
		return fmt.Errorf("page ID is required for update")
	}

	// Use remote version if available, otherwise use local version
	version := page.Version
	if page.RemotePage != nil {
		version = page.RemotePage.Version
	}

	// Update page
	remotePage, err := e.client.UpdatePage(
		page.Confluence.PageID,
		page.Title,
		content,
		version,
		page.Labels,
	)
	if err != nil {
		return fmt.Errorf("failed to update page: %w", err)
	}

	// Update local metadata
	page.Version = remotePage.Version
	page.LastSync = time.Now()

	// Write updated metadata back to file
	if err := e.writer.UpdatePageMetadata(page); err != nil {
		return fmt.Errorf("failed to update local metadata: %w", err)
	}

	return nil
}

// sortByDependency sorts changes so parents are applied before children
func (e *Executor) sortByDependency(changes []*models.Change) ([]*models.Change, error) {
	// Build dependency graph
	pageMap := make(map[string]*models.Change)
	for _, change := range changes {
		if change.Page.Confluence.PageID != "" {
			pageMap[change.Page.Confluence.PageID] = change
		}
	}

	// Simple topological sort: put creates before updates
	var creates, updates []*models.Change
	for _, change := range changes {
		if change.Type == models.ChangeTypeCreate {
			creates = append(creates, change)
		} else {
			updates = append(updates, change)
		}
	}

	// Sort creates by parent dependency
	sort.SliceStable(creates, func(i, j int) bool {
		// Pages without parents come first
		if creates[i].Page.Parent == "" {
			return true
		}
		if creates[j].Page.Parent == "" {
			return false
		}
		return false
	})

	// Combine in order: creates first, then updates
	sorted := append(creates, updates...)
	return sorted, nil
}

// isPageID checks if a string looks like a Confluence page ID
func isPageID(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
