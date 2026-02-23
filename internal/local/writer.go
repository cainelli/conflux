package local

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cainelli/conflux/pkg/models"
	"gopkg.in/yaml.v3"
)

// Writer handles writing pages to the local filesystem
type Writer struct {
	pagesDir string
}

// NewWriter creates a new local file writer
func NewWriter(pagesDir string) *Writer {
	return &Writer{
		pagesDir: pagesDir,
	}
}

// WritePage writes a page to the filesystem
func (w *Writer) WritePage(page *models.Page) error {
	// Build frontmatter
	frontmatter, err := buildFrontmatter(page)
	if err != nil {
		return fmt.Errorf("failed to build frontmatter: %w", err)
	}

	// Combine frontmatter and content
	content := fmt.Sprintf("---\n%s---\n\n%s\n", frontmatter, page.Content)

	// Write to file
	if err := os.WriteFile(page.FilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// UpdatePageMetadata updates only the frontmatter metadata of a page file
func (w *Writer) UpdatePageMetadata(page *models.Page) error {
	// Read existing file
	content, err := os.ReadFile(page.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	str := string(content)

	// Find end of frontmatter
	if !strings.HasPrefix(str, "---\n") {
		return fmt.Errorf("file does not have frontmatter")
	}

	endIdx := strings.Index(str[4:], "\n---\n")
	if endIdx == -1 {
		return fmt.Errorf("frontmatter not properly closed")
	}

	// Extract existing content (after frontmatter)
	pageContent := str[endIdx+9:]

	// Build new frontmatter
	frontmatter, err := buildFrontmatter(page)
	if err != nil {
		return fmt.Errorf("failed to build frontmatter: %w", err)
	}

	// Combine new frontmatter with existing content
	newContent := fmt.Sprintf("---\n%s---\n%s", frontmatter, pageContent)

	// Write back to file
	if err := os.WriteFile(page.FilePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GenerateFilePath generates a file path for a page based on its title
func (w *Writer) GenerateFilePath(title, parentPath string) (string, error) {
	// Sanitize title to create filename
	filename := sanitizeFilename(title) + ".md"

	var filePath string
	if parentPath != "" {
		// Place in same directory as parent
		parentDir := filepath.Dir(parentPath)
		filePath = filepath.Join(parentDir, filename)
	} else {
		// Place in root of pages directory
		filePath = filepath.Join(w.pagesDir, filename)
	}

	// Check if file already exists, append number if needed
	if _, err := os.Stat(filePath); err == nil {
		for i := 1; ; i++ {
			base := strings.TrimSuffix(filename, ".md")
			newFilename := fmt.Sprintf("%s-%d.md", base, i)
			if parentPath != "" {
				parentDir := filepath.Dir(parentPath)
				filePath = filepath.Join(parentDir, newFilename)
			} else {
				filePath = filepath.Join(w.pagesDir, newFilename)
			}
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				break
			}
		}
	}

	return filePath, nil
}

// buildFrontmatter converts a Page to YAML frontmatter
func buildFrontmatter(page *models.Page) (string, error) {
	type FrontmatterData struct {
		Title      string                    `yaml:"title"`
		Confluence models.ConfluenceMetadata `yaml:"confluence"`
		Parent     string                    `yaml:"parent,omitempty"`
		Labels     []string                  `yaml:"labels,omitempty"`
		Version    int                       `yaml:"version"`
		LastSync   string                    `yaml:"last_sync,omitempty"`
	}

	data := FrontmatterData{
		Title:      page.Title,
		Confluence: page.Confluence,
		Parent:     page.Parent,
		Labels:     page.Labels,
		Version:    page.Version,
	}

	if !page.LastSync.IsZero() {
		data.LastSync = page.LastSync.Format(time.RFC3339)
	}

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

// sanitizeFilename converts a page title to a safe filename
func sanitizeFilename(title string) string {
	// Convert to lowercase
	filename := strings.ToLower(title)

	// Replace spaces with hyphens
	filename = strings.ReplaceAll(filename, " ", "-")

	// Remove special characters
	var result strings.Builder
	for _, c := range filename {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result.WriteRune(c)
		}
	}

	filename = result.String()

	// Remove consecutive hyphens
	for strings.Contains(filename, "--") {
		filename = strings.ReplaceAll(filename, "--", "-")
	}

	// Trim hyphens from start and end
	filename = strings.Trim(filename, "-")

	// Ensure non-empty
	if filename == "" {
		filename = "untitled"
	}

	return filename
}
