package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/getyourguide/disaster-recovery-plans/internal/config"
	"github.com/getyourguide/disaster-recovery-plans/internal/confluence"
	"github.com/getyourguide/disaster-recovery-plans/internal/executor"
	"github.com/getyourguide/disaster-recovery-plans/internal/local"
	"github.com/getyourguide/disaster-recovery-plans/internal/planner"
	"github.com/spf13/cobra"
)

// NewApplyCommand creates the apply command
func NewApplyCommand(configPath *string) *cobra.Command {
	var autoApprove bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply changes to Confluence",
		Long:  "Apply the planned changes to Confluence pages.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runApply(*configPath, autoApprove)
		},
	}

	cmd.Flags().BoolVarP(&autoApprove, "auto-approve", "y", false, "skip confirmation prompt")

	return cmd
}

func runApply(configPath string, autoApprove bool) error {
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

	// Prompt for confirmation
	if !autoApprove {
		fmt.Print("Do you want to apply these changes? (yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "yes" && response != "y" {
			fmt.Println("Apply cancelled.")
			return nil
		}
	}

	// Execute changes
	fmt.Println("\nApplying changes...")

	writer := local.NewWriter(cfg.Project.PagesDir)
	exec := executor.NewExecutor(client, writer)

	if err := exec.Execute(changes); err != nil {
		return fmt.Errorf("failed to apply changes: %w", err)
	}

	fmt.Printf("\n✓ Successfully applied %d change(s)\n", len(changes))

	return nil
}
