package cli

import (
	"fmt"

	"github.com/cainelli/conflux/internal/config"
	"github.com/cainelli/conflux/internal/local"
	"github.com/spf13/cobra"
)

// NewValidateCommand creates the validate command
func NewValidateCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate local page files",
		Long:  "Check that local page files have valid frontmatter and structure.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(*configPath)
		},
	}
}

func runValidate(configPath string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	// Load local pages
	loader := local.NewLoader(cfg.Project.PagesDir)
	pages, err := loader.LoadAll()
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(pages) == 0 {
		fmt.Println("No pages found in", cfg.Project.PagesDir)
		return nil
	}

	fmt.Printf("✓ Validated %d page(s)\n", len(pages))

	// Additional validation checks
	var warnings []string

	for _, page := range pages {
		// Check for missing page IDs
		if page.Confluence.PageID == "" {
			warnings = append(warnings, fmt.Sprintf("  ⚠ %s: no page_id (page will be created on apply)", page.FilePath))
		}

		// Check for empty content
		if page.Content == "" {
			warnings = append(warnings, fmt.Sprintf("  ⚠ %s: empty content", page.FilePath))
		}
	}

	if len(warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range warnings {
			fmt.Println(warning)
		}
	}

	return nil
}
