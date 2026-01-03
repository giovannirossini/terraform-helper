package github

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles GitHub API interactions.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new GitHub client with default configuration.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// Content represents a file or directory in GitHub's Contents API.
type Content struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

// Tree represents a GitHub tree response.
type Tree struct {
	Tree []TreeItem `json:"tree"`
}

// TreeItem represents an item in a GitHub tree.
type TreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// FetchProviderDocs retrieves the list of all documentation files for a provider.
// docType should be "r" for resources or "d" for data sources.
// cdktfLanguage determines the CDKTF language to use (empty string for regular Terraform docs).
func (c *Client) FetchProviderDocs(provider, docType, cdktfLanguage string) ([]string, error) {
	// Try using the Git Tree API for better performance with large directories
	items, err := c.fetchProviderDocsViaTree(provider, docType, cdktfLanguage)
	if err == nil && len(items) > 0 {
		return items, nil
	}

	// Fallback to contents API
	return c.fetchProviderDocsViaContents(provider, docType, cdktfLanguage)
}

// FetchDocMarkdown fetches the raw markdown content for a specific resource or data source.
func (c *Client) FetchDocMarkdown(provider, docType, itemName, cdktfLanguage string) (string, error) {
	rawURL := c.buildRawGitHubURL(provider, docType, cdktfLanguage)
	url := fmt.Sprintf("%s/%s.html.markdown", rawURL, itemName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "terraform-helper")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch markdown (status %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	return string(body), nil
}
