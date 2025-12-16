package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/giovannirossini/markdown-render/render"
)

type Config struct {
	Provider     string
	SearchTerm   string
	IsResource   bool
	IsDataSource bool
}

func printUsage() {
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

func main() {
	// Manual flag parsing to support flags after positional arguments
	var isResource, isDataSource bool
	var provider, searchTerm string
	
	// Parse arguments manually
	args := os.Args[1:]
	var positionalArgs []string
	
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-r", "--resource":
			isResource = true
		case "-d", "--datasource":
			isDataSource = true
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Error: Unknown flag: %s\n", arg)
				printUsage()
				os.Exit(1)
			}
			positionalArgs = append(positionalArgs, arg)
		}
	}
	
	if len(positionalArgs) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	provider = positionalArgs[0]
	searchTerm = positionalArgs[1]

	config := Config{
		Provider:     provider,
		SearchTerm:   searchTerm,
		IsResource:   isResource,
		IsDataSource: isDataSource,
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
