package main

import (
	"fmt"
	"os"

	"github.com/getyourguide/disaster-recovery-plans/internal/cli"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "conflux",
		Short: "Confluence as Code - Manage Confluence pages declaratively",
		Long: `Conflux is a Terraform-like CLI tool for managing Confluence Cloud pages as code.
It enables declarative management of Confluence pages using local files with a plan/apply workflow.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	var configPath string
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "config file (default: .conflux/config.yaml)")

	// Add commands
	rootCmd.AddCommand(cli.NewVersionCommand(version, commit))
	rootCmd.AddCommand(cli.NewInitCommand())
	rootCmd.AddCommand(cli.NewPlanCommand(&configPath))
	rootCmd.AddCommand(cli.NewApplyCommand(&configPath))
	rootCmd.AddCommand(cli.NewImportCommand(&configPath))
	rootCmd.AddCommand(cli.NewValidateCommand(&configPath))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
