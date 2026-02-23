package cli

import (
	"fmt"

	"github.com/cainelli/conflux/internal/config"
	"github.com/cainelli/conflux/internal/confluence"
	"github.com/cainelli/conflux/internal/local"
	"github.com/cainelli/conflux/internal/planner"
	"github.com/cainelli/conflux/pkg/models"
	"github.com/spf13/cobra"
)

// NewPlanCommand creates the plan command
func NewPlanCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Preview changes to Confluence pages",
		Long:  "Analyze local pages and show what changes would be made to Confluence.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPlan(*configPath)
		},
	}
}

func runPlan(configPath string) error {
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

	// Load local pages
	loader := local.NewLoader(cfg.Project.PagesDir)
	pages, err := loader.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load pages: %w", err)
	}

	if len(pages) == 0 {
		fmt.Println("No pages found in", cfg.Project.PagesDir)
		return nil
	}

	fmt.Printf("Loaded %d page(s) from %s\n\n", len(pages), cfg.Project.PagesDir)

	// Create planner
	p := planner.NewPlanner(client)

	// Generate plan
	fmt.Println("Analyzing changes...")
	changes, err := p.Plan(pages)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Display plan
	if len(changes) == 0 {
		fmt.Println("No changes detected. Everything is up to date.")
		return nil
	}

	fmt.Printf("\nPlan: %d change(s)\n\n", len(changes))

	for _, change := range changes {
		displayChange(change)
	}

	fmt.Println("\nTo apply these changes, run: conflux apply")

	return nil
}

func displayChange(change *models.Change) {
	page := change.Page

	switch change.Type {
	case models.ChangeTypeCreate:
		fmt.Printf("  + CREATE %s\n", page.Title)
		fmt.Printf("    File: %s\n", page.FilePath)
		if page.Parent != "" {
			fmt.Printf("    Parent: %s\n", page.Parent)
		}
		if len(page.Labels) > 0 {
			fmt.Printf("    Labels: %v\n", page.Labels)
		}

	case models.ChangeTypeUpdate:
		fmt.Printf("  ~ UPDATE %s\n", page.Title)
		fmt.Printf("    File: %s\n", page.FilePath)
		fmt.Printf("    Page ID: %s\n", page.Confluence.PageID)

		// Show content diff if available
		if change.Diff != "" {
			fmt.Println("    Content changes:")
			// Truncate diff if too long
			diff := change.Diff
			if len(diff) > 500 {
				diff = diff[:500] + "\n    ... (diff truncated)"
			}
			fmt.Println("    " + diff)
		}

	case models.ChangeTypeDelete:
		fmt.Printf("  - DELETE %s\n", page.Title)
		fmt.Printf("    Page ID: %s\n", page.Confluence.PageID)
	}

	fmt.Println()
}
