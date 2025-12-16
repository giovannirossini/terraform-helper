package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GitHubContent struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type GitHubTree struct {
	Tree []GitHubTreeItem `json:"tree"`
}

type GitHubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// buildGitHubAPIURL constructs the GitHub API URL for the provider docs
func buildGitHubAPIURL(provider, docType string, useCDKTF bool) string {
	if useCDKTF {
		return fmt.Sprintf("https://api.github.com/repos/hashicorp/terraform-provider-%s/contents/website/docs/cdktf/typescript/%s", provider, docType)
	}
	return fmt.Sprintf("https://api.github.com/repos/hashicorp/terraform-provider-%s/contents/website/docs/%s", provider, docType)
}

// buildGitHubTreeAPIURL constructs the GitHub Tree API URL for better performance with large directories
func buildGitHubTreeAPIURL(provider string) string {
	// Use tree API to get all files at once, recursively
	return fmt.Sprintf("https://api.github.com/repos/hashicorp/terraform-provider-%s/git/trees/main?recursive=1", provider)
}

// buildRawGitHubURL constructs the raw GitHub URL for the provider docs
func buildRawGitHubURL(provider, docType string, useCDKTF bool) string {
	if useCDKTF {
		return fmt.Sprintf("https://raw.githubusercontent.com/hashicorp/terraform-provider-%s/refs/heads/main/website/docs/cdktf/typescript/%s", provider, docType)
	}
	return fmt.Sprintf("https://raw.githubusercontent.com/hashicorp/terraform-provider-%s/refs/heads/main/website/docs/%s", provider, docType)
}

// fetchProviderDocs retrieves the list of all documentation files for a provider
// docType should be "r" for resources or "d" for data sources
// useCDKTF determines whether to use CDKTF TypeScript docs or regular Terraform docs
func fetchProviderDocs(provider, docType string, useCDKTF bool) ([]string, error) {
	// Try using the Git Tree API for better performance with large directories
	items, err := fetchProviderDocsViaTree(provider, docType, useCDKTF)
	if err == nil && len(items) > 0 {
		return items, nil
	}

	// Fallback to contents API
	return fetchProviderDocsViaContents(provider, docType, useCDKTF)
}

// fetchProviderDocsViaTree uses GitHub's Tree API to get all files efficiently
func fetchProviderDocsViaTree(provider, docType string, useCDKTF bool) ([]string, error) {
	apiURL := buildGitHubTreeAPIURL(provider)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "terraform-helper")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var tree GitHubTree
	if err := json.NewDecoder(resp.Body).Decode(&tree); err != nil {
		return nil, err
	}

	// Filter for files in the correct directory
	var prefix string
	if useCDKTF {
		prefix = fmt.Sprintf("website/docs/cdktf/typescript/%s/", docType)
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

// fetchProviderDocsViaContents uses GitHub's Contents API (fallback method)
func fetchProviderDocsViaContents(provider, docType string, useCDKTF bool) ([]string, error) {
	apiURL := buildGitHubAPIURL(provider, docType, useCDKTF)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "terraform-helper")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var contents []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
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

// fetchDocMarkdown fetches the raw markdown content for a specific resource or data source
func fetchDocMarkdown(provider, docType, itemName string, useCDKTF bool) (string, error) {
	rawURL := buildRawGitHubURL(provider, docType, useCDKTF)
	url := fmt.Sprintf("%s/%s.html.markdown", rawURL, itemName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "terraform-helper")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch markdown (status %d)", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
