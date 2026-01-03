package config

import "errors"

var (
	// ErrMissingProvider indicates that the provider argument is missing.
	ErrMissingProvider = errors.New("provider is required")
	// ErrMissingSearchTerm indicates that the search term argument is missing.
	ErrMissingSearchTerm = errors.New("search term is required")
	// ErrConflictingFlags indicates that both resource and datasource flags were specified.
	ErrConflictingFlags = errors.New("cannot specify both -r and -d flags")
)
