package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/giovannirossini/markdown-render/render"

	"github.com/giovannirossini/terraform-helper/internal/config"
	"github.com/giovannirossini/terraform-helper/internal/github"
	"github.com/giovannirossini/terraform-helper/internal/matcher"
	"github.com/giovannirossini/terraform-helper/internal/prompt"
	"github.com/giovannirossini/terraform-helper/internal/version"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: terraform-helper <provider> <name> [flags]\n\n")
	fmt.Fprintf(os.Stderr, "Arguments:\n")
	fmt.Fprintf(os.Stderr, "  provider    Provider name (e.g., aws, google, azurerm)\n")
	fmt.Fprintf(os.Stderr, "  name        Resource or data source name (partial matching supported)\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  -r, --resource      Search in resources (default)\n")
	fmt.Fprintf(os.Stderr, "  -d, --datasource    Search in data sources\n")
	fmt.Fprintf(os.Stderr, "  --cdktf [language]  Use CDKTF documentation (default: typescript)\n")
	fmt.Fprintf(os.Stderr, "  -v, --version       Show version information\n")
	fmt.Fprintf(os.Stderr, "  -h, --help          Show this help message\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper aws api_gateway\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper aws api_gateway_deployment -r\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper google compute_instance\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper azurerm virtual_machine -d\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper aws lambda_function --cdktf\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper aws lambda_function --cdktf typescript\n")
	fmt.Fprintf(os.Stderr, "  terraform-helper aws lambda_function --cdktf python\n")
}

func parseArgs() (*config.Config, error) {
	var isResource, isDataSource bool
	var cdktfLanguage string
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
		case "--cdktf":
			// Check if next argument exists and is not a flag
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Check if it looks like a language name (not a positional arg)
				nextArg := args[i+1]
				// If we already have 2 positional args, treat next arg as language
				if len(positionalArgs) >= 2 {
					cdktfLanguage = nextArg
					i++ // Skip next argument since we consumed it
				} else {
					// Default to typescript
					cdktfLanguage = "typescript"
				}
			} else {
				// No argument provided or next arg is a flag, default to typescript
				cdktfLanguage = "typescript"
			}
		case "-v", "--version":
			fmt.Println(version.String())
			os.Exit(0)
		case "-h", "--help":
			printUsage()
			os.Exit(0)
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag: %s", arg)
			}
			positionalArgs = append(positionalArgs, arg)
		}
	}

	if len(positionalArgs) < 2 {
		return nil, fmt.Errorf("missing required arguments")
	}

	provider = positionalArgs[0]
	searchTerm = positionalArgs[1]

	cfg := &config.Config{
		Provider:      provider,
		SearchTerm:    searchTerm,
		IsResource:    isResource,
		IsDataSource:  isDataSource,
		CDKTFLanguage: cdktfLanguage,
	}

	// Default to resource if neither flag is specified
	if !cfg.IsResource && !cfg.IsDataSource {
		cfg.IsResource = true
	}

	return cfg, nil
}

func main() {
	cfg, err := parseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		printUsage()
		os.Exit(1)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config) error {
	// Fetch list of available resources/datasources
	client := github.NewClient()
	items, err := client.FetchProviderDocs(cfg.Provider, cfg.DocType(), cfg.CDKTFLanguage)
	if err != nil {
		return fmt.Errorf("failed to fetch %s for provider '%s': %w", cfg.DocTypeName(), cfg.Provider, err)
	}

	if len(items) == 0 {
		return fmt.Errorf("no %s found for provider '%s' (provider may not exist or use different structure)", cfg.DocTypeName(), cfg.Provider)
	}

	// Find matches
	matches := matcher.FindMatches(cfg.SearchTerm, items)

	if len(matches) == 0 {
		return fmt.Errorf("no %s found matching '%s'", cfg.DocTypeName(), cfg.SearchTerm)
	}

	var selectedItem string

	// Check for exact match
	exactMatch := matcher.FindExactMatch(cfg.SearchTerm, matches)
	if exactMatch != "" {
		selectedItem = exactMatch
	} else if len(matches) == 1 {
		// Only one match, use it
		selectedItem = matches[0]
	} else {
		// Multiple matches, show selection
		selected, err := prompt.Select(matches)
		if err != nil {
			return fmt.Errorf("selection failed: %w", err)
		}
		selectedItem = selected
	}

	// Fetch and display the markdown
	markdown, err := client.FetchDocMarkdown(cfg.Provider, cfg.DocType(), selectedItem, cfg.CDKTFLanguage)
	if err != nil {
		return fmt.Errorf("failed to fetch markdown: %w", err)
	}

	// Render the markdown with colors and formatting
	render.Render(markdown)
	return nil
}
