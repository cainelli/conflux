package models

import "time"

// Page represents a Confluence page with both local and remote metadata
type Page struct {
	// Local metadata (from frontmatter)
	Title       string            `yaml:"title"`
	Confluence  ConfluenceMetadata `yaml:"confluence"`
	Parent      string            `yaml:"parent,omitempty"` // Relative path or page_id
	Labels      []string          `yaml:"labels,omitempty"`
	Version     int               `yaml:"version"`
	LastSync    time.Time         `yaml:"last_sync,omitempty"`

	// Content
	Content     string            // Raw content (Markdown or Confluence Storage Format)
	ContentType ContentType       // Detected format

	// Local file information
	FilePath    string            // Path to local file

	// Remote state (fetched from Confluence)
	RemotePage  *RemotePage       // nil if page doesn't exist remotely
}

// ConfluenceMetadata contains Confluence-specific metadata
type ConfluenceMetadata struct {
	PageID   string `yaml:"page_id,omitempty"`   // Set after creation/import
	SpaceKey string `yaml:"space_key"`
}

// RemotePage represents the current state of a page in Confluence
type RemotePage struct {
	ID       string
	Title    string
	SpaceKey string
	Version  int
	Content  string // Confluence Storage Format
	Labels   []string
	ParentID string
}

// ContentType represents the format of page content
type ContentType string

const (
	ContentTypeConfluence ContentType = "confluence"
)

// ChangeType represents the type of change detected
type ChangeType string

const (
	ChangeTypeCreate ChangeType = "CREATE"
	ChangeTypeUpdate ChangeType = "UPDATE"
	ChangeTypeDelete ChangeType = "DELETE"
	ChangeTypeNone   ChangeType = "NONE"
)

// Change represents a detected change between local and remote state
type Change struct {
	Type         ChangeType
	Page         *Page
	LocalContent string
	RemoteContent string
	Diff         string // Human-readable diff
}
