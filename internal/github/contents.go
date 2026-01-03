package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// fetchProviderDocsViaContents uses GitHub's Contents API (fallback method).
func (c *Client) fetchProviderDocsViaContents(provider, docType, cdktfLanguage string) ([]string, error) {
	apiURL := c.buildGitHubAPIURL(provider, docType, cdktfLanguage)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", "terraform-helper")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var contents []Content
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var items []string
	for _, content := range contents {
		if content.Type == "file" && strings.HasSuffix(content.Name, ".html.markdown") {
			// Remove the .html.markdown extension
			itemName := strings.TrimSuffix(content.Name, ".html.markdown")
			items = append(items, itemName)
		}
	}

	return items, nil
}

// buildGitHubAPIURL constructs the GitHub API URL for the provider docs.
func (c *Client) buildGitHubAPIURL(provider, docType, cdktfLanguage string) string {
	if cdktfLanguage != "" {
		return fmt.Sprintf("%s/repos/hashicorp/terraform-provider-%s/contents/website/docs/cdktf/%s/%s", c.baseURL, provider, cdktfLanguage, docType)
	}
	return fmt.Sprintf("%s/repos/hashicorp/terraform-provider-%s/contents/website/docs/%s", c.baseURL, provider, docType)
}

// buildRawGitHubURL constructs the raw GitHub URL for the provider docs.
func (c *Client) buildRawGitHubURL(provider, docType, cdktfLanguage string) string {
	if cdktfLanguage != "" {
		return fmt.Sprintf("https://raw.githubusercontent.com/hashicorp/terraform-provider-%s/refs/heads/main/website/docs/cdktf/%s/%s", provider, cdktfLanguage, docType)
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/hashicorp/terraform-provider-%s/refs/heads/main/website/docs/%s", provider, docType)
}
