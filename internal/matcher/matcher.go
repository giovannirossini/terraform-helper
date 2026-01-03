package matcher

import "strings"

// FindMatches finds all resources that contain the search term.
func FindMatches(searchTerm string, resources []string) []string {
	searchLower := strings.ToLower(searchTerm)
	var matches []string

	for _, resource := range resources {
		resourceLower := strings.ToLower(resource)
		if strings.Contains(resourceLower, searchLower) {
			matches = append(matches, resource)
		}
	}

	return matches
}

// FindExactMatch checks if there's an exact match in the list.
func FindExactMatch(searchTerm string, matches []string) string {
	searchLower := strings.ToLower(searchTerm)

	for _, match := range matches {
		if strings.ToLower(match) == searchLower {
			return match
		}
	}

	return ""
}
