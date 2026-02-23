package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Confluence ConfluenceConfig `mapstructure:"confluence"`
	Project    ProjectConfig    `mapstructure:"project"`
	Behavior   BehaviorConfig   `mapstructure:"behavior"`
}

// ConfluenceConfig contains Confluence connection settings
type ConfluenceConfig struct {
	BaseURL  string     `mapstructure:"base_url"`
	SpaceKey string     `mapstructure:"space_key"`
	Auth     AuthConfig `mapstructure:"auth"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Type  string `mapstructure:"type"`  // "token" or "basic"
	Token string `mapstructure:"token"` // Can be env var reference like ${CONFLUX_API_TOKEN}
	Email string `mapstructure:"email,omitempty"`
}

// ProjectConfig contains project-specific settings
type ProjectConfig struct {
	PagesDir string `mapstructure:"pages_dir"`
}

// BehaviorConfig contains tool behavior settings
type BehaviorConfig struct {
	CreateMissingParents bool `mapstructure:"create_missing_parents"`
	UpdateLabels         bool `mapstructure:"update_labels"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Project: ProjectConfig{
			PagesDir: "pages",
		},
		Behavior: BehaviorConfig{
			CreateMissingParents: true,
			UpdateLabels:         true,
		},
	}
}

// Load loads configuration from .conflux/config.yaml
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	cfg := DefaultConfig()
	v.SetDefault("project.pages_dir", cfg.Project.PagesDir)
	v.SetDefault("behavior.create_missing_parents", cfg.Behavior.CreateMissingParents)
	v.SetDefault("behavior.update_labels", cfg.Behavior.UpdateLabels)

	// Configure viper
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for .conflux/config.yaml in current directory
		v.AddConfigPath(".conflux")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	// Enable environment variable substitution
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: run 'conflux init' to create one")
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Unmarshal into config struct
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Expand environment variables in token
	if cfg.Confluence.Auth.Token != "" {
		cfg.Confluence.Auth.Token = os.ExpandEnv(cfg.Confluence.Auth.Token)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	if c.Confluence.BaseURL == "" {
		return fmt.Errorf("confluence.base_url is required")
	}
	if c.Confluence.SpaceKey == "" {
		return fmt.Errorf("confluence.space_key is required")
	}
	if c.Confluence.Auth.Type == "" {
		return fmt.Errorf("confluence.auth.type is required")
	}
	if c.Confluence.Auth.Type == "token" && c.Confluence.Auth.Token == "" {
		return fmt.Errorf("confluence.auth.token is required when auth.type is 'token'")
	}
	return nil
}

