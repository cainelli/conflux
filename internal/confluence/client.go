package confluence

import (
	"fmt"
	"regexp"

	"github.com/getyourguide/disaster-recovery-plans/internal/config"
	"github.com/getyourguide/disaster-recovery-plans/pkg/models"
	"github.com/go-resty/resty/v2"
)

var (
	reACLocalID = regexp.MustCompile(`\s+ac:local-id="[^"]*"`)
	reLocalID   = regexp.MustCompile(`\s+local-id="[^"]*"`)
)

// cleanupContent removes ephemeral Confluence attributes (like local-id)
// that change between fetches and are not meaningful for content comparison.
func cleanupContent(content string) string {
	content = reACLocalID.ReplaceAllString(content, "")
	content = reLocalID.ReplaceAllString(content, "")
	return content
}

// Client wraps the Confluence Cloud REST API v2
type Client struct {
	baseURL string
	client  *resty.Client
}

// NewClient creates a new Confluence API client
func NewClient(cfg *config.Config) (*Client, error) {
	client := resty.New()
	client.SetBaseURL(cfg.Confluence.BaseURL)

	// Set authentication
	// Confluence Cloud REST API v2 requires Basic Auth (email + API token)
	switch cfg.Confluence.Auth.Type {
	case "token":
		// For token auth, we need an email address for Basic Auth
		if cfg.Confluence.Auth.Email == "" {
			return nil, fmt.Errorf("email is required for token authentication with Confluence Cloud")
		}
		client.SetBasicAuth(cfg.Confluence.Auth.Email, cfg.Confluence.Auth.Token)
	case "basic":
		client.SetBasicAuth(cfg.Confluence.Auth.Email, cfg.Confluence.Auth.Token)
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.Confluence.Auth.Type)
	}

	// Set common headers
	client.SetHeader("Accept", "application/json")
	client.SetHeader("Content-Type", "application/json")

	// Enable retries
	client.SetRetryCount(3)

	return &Client{
		baseURL: cfg.Confluence.BaseURL,
		client:  client,
	}, nil
}

// GetPage fetches a page by ID from Confluence using REST API v2
func (c *Client) GetPage(pageID string) (*models.RemotePage, error) {
	type response struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		SpaceID string `json:"spaceId"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
		Body struct {
			Storage struct {
				Representation string `json:"representation"`
				Value          string `json:"value"`
			} `json:"storage"`
		} `json:"body"`
		ParentID string `json:"parentId"`
		Labels   struct {
			Results []struct {
				Name string `json:"name"`
			} `json:"results"`
		} `json:"labels"`
	}

	resp, err := c.client.R().
		SetPathParam("pageId", pageID).
		SetQueryParam("body-format", "storage").
		SetQueryParam("include-labels", "true").
		SetResult(&response{}).
		Get("/api/v2/pages/{pageId}")

	if err != nil {
		return nil, fmt.Errorf("failed to get page: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*response)

	// Extract labels
	labels := make([]string, 0, len(result.Labels.Results))
	for _, label := range result.Labels.Results {
		labels = append(labels, label.Name)
	}

	return &models.RemotePage{
		ID:       result.ID,
		Title:    result.Title,
		SpaceKey: result.SpaceID,
		Version:  result.Version.Number,
		Content:  cleanupContent(result.Body.Storage.Value),
		Labels:   labels,
		ParentID: result.ParentID,
	}, nil
}

// GetSpaceID converts a space key to a space ID
func (c *Client) GetSpaceID(spaceKey string) (string, error) {
	type response struct {
		Results []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		} `json:"results"`
	}

	resp, err := c.client.R().
		SetQueryParam("keys", spaceKey).
		SetResult(&response{}).
		Get("/api/v2/spaces")

	if err != nil {
		return "", fmt.Errorf("failed to get space: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("API error: %s - %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*response)
	if len(result.Results) == 0 {
		return "", fmt.Errorf("space not found: %s", spaceKey)
	}

	return result.Results[0].ID, nil
}

// CreatePage creates a new page in Confluence using REST API v2
func (c *Client) CreatePage(spaceKey, title, content string, parentID string, labels []string) (*models.RemotePage, error) {
	// Convert space key to space ID
	spaceID, err := c.GetSpaceID(spaceKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get space ID: %w", err)
	}

	type request struct {
		SpaceID  string `json:"spaceId"`
		Status   string `json:"status"`
		Title    string `json:"title"`
		ParentID string `json:"parentId,omitempty"`
		Body     struct {
			Representation string `json:"representation"`
			Value          string `json:"value"`
		} `json:"body"`
	}

	req := request{
		SpaceID:  spaceID,
		Status:   "current",
		Title:    title,
		ParentID: parentID,
	}
	req.Body.Representation = "storage"
	req.Body.Value = content

	type response struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
	}

	resp, err := c.client.R().
		SetBody(req).
		SetResult(&response{}).
		Post("/api/v2/pages")

	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*response)

	// Add labels if provided
	if len(labels) > 0 {
		if err := c.AddLabels(result.ID, labels); err != nil {
			// Log warning but don't fail the create operation
			fmt.Printf("Warning: failed to add labels: %v\n", err)
		}
	}

	return &models.RemotePage{
		ID:       result.ID,
		Title:    result.Title,
		SpaceKey: spaceKey,
		Version:  result.Version.Number,
		Content:  content,
		Labels:   labels,
		ParentID: parentID,
	}, nil
}

// UpdatePage updates an existing page in Confluence using REST API v2
func (c *Client) UpdatePage(pageID, title, content string, version int, labels []string) (*models.RemotePage, error) {
	type request struct {
		ID      string `json:"id"`
		Status  string `json:"status"`
		Title   string `json:"title"`
		Body    struct {
			Representation string `json:"representation"`
			Value          string `json:"value"`
		} `json:"body"`
		Version struct {
			Number  int    `json:"number"`
			Message string `json:"message,omitempty"`
		} `json:"version"`
	}

	req := request{
		ID:     pageID,
		Status: "current",
		Title:  title,
	}
	req.Body.Representation = "storage"
	req.Body.Value = content
	req.Version.Number = version + 1
	req.Version.Message = "Updated via Conflux"

	type response struct {
		ID      string `json:"id"`
		Title   string `json:"title"`
		Version struct {
			Number int `json:"number"`
		} `json:"version"`
	}

	resp, err := c.client.R().
		SetPathParam("pageId", pageID).
		SetBody(req).
		SetResult(&response{}).
		Put("/api/v2/pages/{pageId}")

	if err != nil {
		return nil, fmt.Errorf("failed to update page: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status(), string(resp.Body()))
	}

	result := resp.Result().(*response)

	// Update labels if provided
	if len(labels) > 0 {
		if err := c.SetLabels(result.ID, labels); err != nil {
			fmt.Printf("Warning: failed to update labels: %v\n", err)
		}
	}

	return &models.RemotePage{
		ID:      result.ID,
		Title:   result.Title,
		Version: result.Version.Number,
		Content: content,
		Labels:  labels,
	}, nil
}

// AddLabels adds labels to a page using REST API v2
func (c *Client) AddLabels(pageID string, labels []string) error {
	for _, label := range labels {
		type request struct {
			Prefix string `json:"prefix"`
			Name   string `json:"name"`
		}

		req := request{
			Prefix: "global",
			Name:   label,
		}

		resp, err := c.client.R().
			SetPathParam("pageId", pageID).
			SetBody(req).
			Post("/api/v2/pages/{pageId}/labels")

		if err != nil {
			return fmt.Errorf("failed to add label %s: %w", label, err)
		}

		if resp.IsError() {
			return fmt.Errorf("API error adding label %s: %s", label, resp.Status())
		}
	}

	return nil
}

// SetLabels sets the exact label set for a page (removes old labels, adds new ones)
func (c *Client) SetLabels(pageID string, labels []string) error {
	// Get current labels
	page, err := c.GetPage(pageID)
	if err != nil {
		return err
	}

	// Remove labels that are not in the new set
	for _, oldLabel := range page.Labels {
		found := false
		for _, newLabel := range labels {
			if oldLabel == newLabel {
				found = true
				break
			}
		}
		if !found {
			if err := c.RemoveLabel(pageID, oldLabel); err != nil {
				return err
			}
		}
	}

	// Add new labels
	for _, label := range labels {
		found := false
		for _, oldLabel := range page.Labels {
			if label == oldLabel {
				found = true
				break
			}
		}
		if !found {
			if err := c.AddLabels(pageID, []string{label}); err != nil {
				return err
			}
		}
	}

	return nil
}

// RemoveLabel removes a label from a page using REST API v2
func (c *Client) RemoveLabel(pageID, label string) error {
	resp, err := c.client.R().
		SetPathParam("pageId", pageID).
		SetPathParam("labelId", label).
		Delete("/api/v2/pages/{pageId}/labels/{labelId}")

	if err != nil {
		return fmt.Errorf("failed to remove label %s: %w", label, err)
	}

	if resp.IsError() && resp.StatusCode() != 404 {
		return fmt.Errorf("API error removing label %s: %s", label, resp.Status())
	}

	return nil
}

// GetChildPages fetches all child pages of a parent page using REST API v2
func (c *Client) GetChildPages(pageID string) ([]*models.RemotePage, error) {
	type response struct {
		Results []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"results"`
	}

	resp, err := c.client.R().
		SetPathParam("pageId", pageID).
		SetResult(&response{}).
		Get("/api/v2/pages/{pageId}/children")

	if err != nil {
		return nil, fmt.Errorf("failed to get child pages: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API error: %s", resp.Status())
	}

	result := resp.Result().(*response)
	children := make([]*models.RemotePage, 0, len(result.Results))

	// Fetch full page details for each child
	for _, child := range result.Results {
		page, err := c.GetPage(child.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get child page %s: %w", child.ID, err)
		}
		children = append(children, page)
	}

	return children, nil
}

// TestConnection tests the connection to Confluence using REST API v2
func (c *Client) TestConnection() error {
	resp, err := c.client.R().
		Get("/api/v2/spaces")

	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("authentication failed: %s", resp.Status())
	}

	return nil
}

