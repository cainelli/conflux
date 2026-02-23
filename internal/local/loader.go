package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cainelli/conflux/pkg/models"
	"gopkg.in/yaml.v3"
)

// Loader handles loading pages from the local filesystem
type Loader struct {
	pagesDir string
}

// NewLoader creates a new local file loader
func NewLoader(pagesDir string) *Loader {
	return &Loader{
		pagesDir: pagesDir,
	}
}

// LoadAll loads all pages from the pages directory
func (l *Loader) LoadAll() ([]*models.Page, error) {
	var pages []*models.Page

	err := filepath.Walk(l.pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		page, err := l.LoadPage(path)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return pages, nil
}

// LoadPage loads a single page from a file
func (l *Loader) LoadPage(filePath string) (*models.Page, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse frontmatter and content
	page, pageContent, err := parseFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set file path
	page.FilePath = filePath

	// Set content
	page.Content = pageContent

	// Detect content type
	page.ContentType = models.ContentTypeConfluence

	return page, nil
}

// parseFrontmatter parses YAML frontmatter from file content
func parseFrontmatter(content []byte) (*models.Page, string, error) {
	str := string(content)

	// Check for frontmatter
	if !strings.HasPrefix(str, "---\n") {
		return nil, "", fmt.Errorf("file must start with YAML frontmatter (---)")
	}

	// Find end of frontmatter
	endIdx := strings.Index(str[4:], "\n---\n")
	if endIdx == -1 {
		return nil, "", fmt.Errorf("frontmatter not properly closed (missing closing ---)")
	}

	// Extract frontmatter and content
	frontmatterStr := str[4 : endIdx+4]
	pageContent := strings.TrimSpace(str[endIdx+9:])

	// Parse frontmatter YAML
	var page models.Page
	if err := yaml.Unmarshal([]byte(frontmatterStr), &page); err != nil {
		return nil, "", fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	// Validate required fields
	if page.Title == "" {
		return nil, "", fmt.Errorf("title is required in frontmatter")
	}
	if page.Confluence.SpaceKey == "" {
		return nil, "", fmt.Errorf("confluence.space_key is required in frontmatter")
	}

	return &page, pageContent, nil
}

