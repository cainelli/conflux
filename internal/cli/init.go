package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cainelli/conflux/internal/config"
	"github.com/cainelli/conflux/internal/confluence"
	"github.com/spf13/cobra"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new conflux project",
		Long:  "Create a new .conflux directory with configuration file and set up project structure.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing configuration")

	return cmd
}

func runInit(force bool) error {
	// Check if .conflux directory already exists
	if _, err := os.Stat(".conflux"); err == nil && !force {
		return fmt.Errorf(".conflux directory already exists. Use --force to overwrite")
	}

	fmt.Println("Initializing Conflux project...")
	fmt.Println()

	// Interactive prompts
	reader := bufio.NewReader(os.Stdin)

	// Confluence URL
	fmt.Print("Confluence base URL (e.g., https://your-domain.atlassian.net/wiki): ")
	baseURL, _ := reader.ReadString('\n')
	baseURL = strings.TrimSpace(baseURL)

	// Space key
	fmt.Print("Default space key (e.g., MYSPACE): ")
	spaceKey, _ := reader.ReadString('\n')
	spaceKey = strings.TrimSpace(spaceKey)

	// Email (required for Confluence Cloud)
	fmt.Print("Email address (required for authentication): ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	// Auth type
	fmt.Print("Authentication type [token/basic] (default: token): ")
	authType, _ := reader.ReadString('\n')
	authType = strings.TrimSpace(authType)
	if authType == "" {
		authType = "token"
	}

	// API token
	var authToken string
	if authType == "basic" || authType == "token" {
		fmt.Println("API token will be read from CONFLUENCE_API_TOKEN environment variable")
		authToken = "${CONFLUENCE_API_TOKEN}"
	}

	// Create .conflux directory
	if err := os.MkdirAll(".conflux", 0755); err != nil {
		return fmt.Errorf("failed to create .conflux directory: %w", err)
	}

	// Create config.yaml
	configContent := fmt.Sprintf(`confluence:
  base_url: %s
  space_key: %s
  auth:
    type: %s
    token: %s
`, baseURL, spaceKey, authType, authToken)

	if email != "" {
		configContent += fmt.Sprintf("    email: %s\n", email)
	}

	configContent += `
project:
  pages_dir: pages

behavior:
  create_missing_parents: true
  update_labels: true
`

	configPath := filepath.Join(".conflux", "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create .gitignore
	gitignorePath := filepath.Join(".conflux", ".gitignore")
	gitignoreContent := `# Never commit API tokens
config.local.yaml
*.secret
`
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	// Create pages directory
	if err := os.MkdirAll("pages", 0755); err != nil {
		return fmt.Errorf("failed to create pages directory: %w", err)
	}

	// Create example page
	examplePage := `---
title: "Getting Started"
confluence:
  space_key: ` + spaceKey + `
labels:
  - documentation
---

# Getting Started with Conflux

This is an example page managed by Conflux.

## How to use

1. Edit this file or create new .md files in the pages/ directory
2. Run 'conflux plan' to preview changes
3. Run 'conflux apply' to publish changes to Confluence
`

	examplePath := filepath.Join("pages", "index.md")
	if err := os.WriteFile(examplePath, []byte(examplePage), 0644); err != nil {
		return fmt.Errorf("failed to write example page: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Configuration created at .conflux/config.yaml")
	fmt.Println("✓ Example page created at pages/index.md")

	// Test connection if token is not an env var reference
	if !strings.HasPrefix(authToken, "${") {
		fmt.Println()
		fmt.Println("Testing connection to Confluence...")

		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("⚠ Warning: Failed to load config: %v\n", err)
			return nil
		}

		client, err := confluence.NewClient(cfg)
		if err != nil {
			fmt.Printf("⚠ Warning: Failed to create client: %v\n", err)
			return nil
		}

		if err := client.TestConnection(); err != nil {
			fmt.Printf("⚠ Warning: Connection test failed: %v\n", err)
			fmt.Println("Please verify your configuration and credentials.")
			return nil
		}

		fmt.Println("✓ Successfully connected to Confluence")
	} else {
		fmt.Println()
		fmt.Println("⚠ Remember to set the CONFLUENCE_API_TOKEN environment variable")
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Set your API token: export CONFLUENCE_API_TOKEN=your_token_here")
	fmt.Println("  2. Edit pages in the pages/ directory")
	fmt.Println("  3. Run 'conflux plan' to preview changes")
	fmt.Println("  4. Run 'conflux apply' to publish to Confluence")

	return nil
}
