package config

// Config holds the application configuration.
type Config struct {
	Provider      string
	SearchTerm    string
	IsResource    bool
	IsDataSource  bool
	CDKTFLanguage string
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	if c.Provider == "" {
		return ErrMissingProvider
	}
	if c.SearchTerm == "" {
		return ErrMissingSearchTerm
	}
	if c.IsResource && c.IsDataSource {
		return ErrConflictingFlags
	}
	return nil
}

// DocType returns the document type string ("r" for resources, "d" for data sources).
func (c *Config) DocType() string {
	if c.IsDataSource {
		return "d"
	}
	return "r"
}

// DocTypeName returns the human-readable document type name.
func (c *Config) DocTypeName() string {
	if c.IsDataSource {
		return "data sources"
	}
	return "resources"
}
