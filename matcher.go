package main

import (
	"strings"
)

// findMatches finds all resources that contain the search term
func findMatches(searchTerm string, resources []string) []string {
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

// findExactMatch checks if there's an exact match in the list
func findExactMatch(searchTerm string, matches []string) string {
	searchLower := strings.ToLower(searchTerm)

	for _, match := range matches {
		if strings.ToLower(match) == searchLower {
			return match
		}
	}

	return ""
}
