package config

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigSource allows addons to provide config from remote sources.
// This is duplicated from addons package to avoid circular imports.
type ConfigSource interface {
	Load() ([]byte, error)
	Watch() <-chan struct{}
}

// Load loads either a single file or a directory containing multiple config files.
// After loading, it validates the configuration for duplicates, unknown references, etc.
func Load(path string) (*Config, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles:     make(map[string]Role),
		Subjects:  make(map[string]Subject),
		Policies:  []Policy{},
	}

	if info.IsDir() {
		// Walk directory and load all YAML files
		err := filepath.WalkDir(path, func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				return nil
			}

			// Only load *.yml, *.yaml, or *.aegis files
			ext := strings.ToLower(filepath.Ext(p))
			if ext != ".yml" && ext != ".yaml" && ext != ".aegis" {
				return nil
			}

			// Parse and merge into cfg
			if err := loadAndMerge(cfg, p); err != nil {
				return fmt.Errorf("failed to load %s: %w", p, err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Single file
		if err := loadAndMerge(cfg, path); err != nil {
			return nil, fmt.Errorf("failed to load file: %w", err)
		}
	}

	// Validate the entire configuration
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return cfg, nil
}

// LoadFromSource loads config from a ConfigSource (provided by addons).
// The source returns raw YAML bytes (S3, GitHub, HTTP, etc.).
func LoadFromSource(source ConfigSource) (*Config, error) {
	data, err := source.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load from source: %w", err)
	}

	cfg := &Config{
		Resources: make(map[string]Resource),
		Roles:     make(map[string]Role),
		Subjects:  make(map[string]Subject),
		Policies:  []Policy{},
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return cfg, nil
}

// loadAndMerge parses a single YAML file and merges it into the existing config.
func loadAndMerge(cfg *Config, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var partial Config
	if err := yaml.Unmarshal(data, &partial); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Merge resources
	for key, resource := range partial.Resources {
		if _, exists := cfg.Resources[key]; exists {
			return fmt.Errorf("duplicate resource key: %s", key)
		}
		cfg.Resources[key] = resource
	}

	// Merge roles
	for key, role := range partial.Roles {
		if _, exists := cfg.Roles[key]; exists {
			return fmt.Errorf("duplicate role key: %s", key)
		}
		cfg.Roles[key] = role
	}

	// Merge subjects
	for key, subject := range partial.Subjects {
		if _, exists := cfg.Subjects[key]; exists {
			return fmt.Errorf("duplicate subject key: %s", key)
		}
		cfg.Subjects[key] = subject
	}

	// Append policies (policies can be duplicated across files)
	cfg.Policies = append(cfg.Policies, partial.Policies...)

	return nil
}
