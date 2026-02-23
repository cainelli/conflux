package cli

import (
	"fmt"
	"time"

	"github.com/getyourguide/disaster-recovery-plans/internal/config"
	"github.com/getyourguide/disaster-recovery-plans/internal/confluence"
	"github.com/getyourguide/disaster-recovery-plans/internal/local"
	"github.com/getyourguide/disaster-recovery-plans/pkg/models"
	"github.com/spf13/cobra"
)

// NewImportCommand creates the import command
func NewImportCommand(configPath *string) *cobra.Command {
	var pageID string
	var recursive bool

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import existing Confluence pages",
		Long:  "Import pages from Confluence to local files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageID == "" {
				return fmt.Errorf("--page-id is required")
			}
			return runImport(*configPath, pageID, recursive)
		},
	}

	cmd.Flags().StringVarP(&pageID, "page-id", "p", "", "Confluence page ID to import (required)")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "recursively import child pages")
	cmd.MarkFlagRequired("page-id")

	return cmd
}

func runImport(configPath, pageID string, recursive bool) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Create Confluence client
	client, err := confluence.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create writer
	writer := local.NewWriter(cfg.Project.PagesDir)

	// Import page
	fmt.Printf("Importing page %s...\n", pageID)

	count := 0
	if err := importPage(client, writer, cfg, pageID, "", recursive, &count); err != nil {
		return fmt.Errorf("failed to import page: %w", err)
	}

	fmt.Printf("\n✓ Successfully imported %d page(s)\n", count)

	return nil
}

func importPage(client *confluence.Client, writer *local.Writer, cfg *config.Config, pageID, parentPath string, recursive bool, count *int) error {
	// Fetch page from Confluence
	remotePage, err := client.GetPage(pageID)
	if err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}

	content := remotePage.Content

	// Generate file path
	filePath, err := writer.GenerateFilePath(remotePage.Title, parentPath)
	if err != nil {
		return fmt.Errorf("failed to generate file path: %w", err)
	}

	// Create page object
	// Use the space ID from the remote page (numeric ID)
	page := &models.Page{
		Title: remotePage.Title,
		Confluence: models.ConfluenceMetadata{
			PageID:   remotePage.ID,
			SpaceKey: remotePage.SpaceKey, // This is actually the space ID from API v2
		},
		Labels:      remotePage.Labels,
		Version:     remotePage.Version,
		LastSync:    time.Now(),
		Content:     content,
		ContentType: models.ContentTypeConfluence,
		FilePath:    filePath,
		RemotePage:  remotePage,
	}

	// Set parent if available
	if remotePage.ParentID != "" {
		page.Parent = remotePage.ParentID
	}

	// Write page to file
	if err := writer.WritePage(page); err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}

	fmt.Printf("  ✓ Imported: %s -> %s\n", remotePage.Title, filePath)
	*count++

	// Import children if recursive
	if recursive {
		children, err := client.GetChildPages(pageID)
		if err != nil {
			return fmt.Errorf("failed to get child pages: %w", err)
		}

		for _, child := range children {
			if err := importPage(client, writer, cfg, child.ID, filePath, recursive, count); err != nil {
				return err
			}
		}
	}

	return nil
}


