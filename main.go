package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/giovannirossini/markdown-render/render"
)

type Config struct {
	Provider     string
	SearchTerm   string
	IsResource   bool
	IsDataSource bool
}

func main() {
	// Define flags
	resourceFlag := flag.Bool("r", false, "Search in resources (default)")
	resourceFlagLong := flag.Bool("resource", false, "Search in resources (default)")
	datasourceFlag := flag.Bool("d", false, "Search in data sources")
	datasourceFlagLong := flag.Bool("datasource", false, "Search in data sources")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: terraform-helper <provider> <name> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  provider    Provider name (e.g., aws, google, azurerm)\n")
		fmt.Fprintf(os.Stderr, "  name        Resource or data source name (partial matching supported)\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fmt.Fprintf(os.Stderr, "  -r, --resource      Search in resources (default)\n")
		fmt.Fprintf(os.Stderr, "  -d, --datasource    Search in data sources\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  terraform-helper aws api_gateway\n")
		fmt.Fprintf(os.Stderr, "  terraform-helper aws api_gateway_deployment -r\n")
		fmt.Fprintf(os.Stderr, "  terraform-helper google compute_instance\n")
		fmt.Fprintf(os.Stderr, "  terraform-helper azurerm virtual_machine -d\n")
	}

	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	config := Config{
		Provider:     args[0],
		SearchTerm:   args[1],
		IsResource:   *resourceFlag || *resourceFlagLong,
		IsDataSource: *datasourceFlag || *datasourceFlagLong,
	}

	// Default to resource if neither flag is specified
	if !config.IsResource && !config.IsDataSource {
		config.IsResource = true
	}

	// Can't specify both
	if config.IsResource && config.IsDataSource {
		fmt.Fprintln(os.Stderr, "Error: Cannot specify both -r and -d flags")
		os.Exit(1)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(config Config) error {
	docType := "r"
	docTypeName := "resources"
	if config.IsDataSource {
		docType = "d"
		docTypeName = "data sources"
	}

	// Fetch list of available resources/datasources
	items, err := fetchProviderDocs(config.Provider, docType)
	if err != nil {
		return fmt.Errorf("failed to fetch %s for provider '%s': %w", docTypeName, config.Provider, err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no %s found for provider '%s' (provider may not exist or use different structure)", docTypeName, config.Provider)
	}

	// Find matches
	matches := findMatches(config.SearchTerm, items)

	if len(matches) == 0 {
		return fmt.Errorf("no %s found matching '%s'", docTypeName, config.SearchTerm)
	}

	var selectedItem string

	// Check for exact match
	exactMatch := findExactMatch(config.SearchTerm, matches)
	if exactMatch != "" {
		selectedItem = exactMatch
	} else if len(matches) == 1 {
		// Only one match, use it
		selectedItem = matches[0]
	} else {
		// Multiple matches, show selection
		selected, err := promptUserSelection(matches)
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}
		selectedItem = selected
	}

	// Fetch and display the markdown
	markdown, err := fetchDocMarkdown(config.Provider, docType, selectedItem)
	if err != nil {
		return fmt.Errorf("failed to fetch markdown: %w", err)
	}

	// Render the markdown with colors and formatting
	render.Render(markdown)
	return nil
}
