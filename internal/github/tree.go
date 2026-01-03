package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// fetchProviderDocsViaTree uses GitHub's Tree API to get all files efficiently.
func (c *Client) fetchProviderDocsViaTree(provider, docType, cdktfLanguage string) ([]string, error) {
	apiURL := c.buildGitHubTreeAPIURL(provider)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "terraform-helper")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var tree Tree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Filter for files in the correct directory
	var prefix string
	if cdktfLanguage != "" {
		prefix = fmt.Sprintf("website/docs/cdktf/%s/%s/", cdktfLanguage, docType)
	} else {
		prefix = fmt.Sprintf("website/docs/%s/", docType)
	}
	var items []string

	for _, item := range tree.Tree {
		if item.Type == "blob" && strings.HasPrefix(item.Path, prefix) && strings.HasSuffix(item.Path, ".html.markdown") {
			// Extract just the filename without directory and extension
			filename := strings.TrimPrefix(item.Path, prefix)
			itemName := strings.TrimSuffix(filename, ".html.markdown")
			// Only add if it's directly in the directory (no subdirectories)
			if !strings.Contains(itemName, "/") {
				items = append(items, itemName)
			}
		}
	}

	return items, nil
}

// buildGitHubTreeAPIURL constructs the GitHub Tree API URL for better performance with large directories.
func (c *Client) buildGitHubTreeAPIURL(provider string) string {
	// Use tree API to get all files at once, recursively
	return fmt.Sprintf("%s/repos/hashicorp/terraform-provider-%s/git/trees/main?recursive=1", c.baseURL, provider)
}
