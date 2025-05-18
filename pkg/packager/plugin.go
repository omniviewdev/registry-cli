package packager

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type PluginMetadata struct {
	ID           string       `yaml:"id"`
	Version      string       `yaml:"version"`
	Name         string       `yaml:"name"`
	Icon         string       `yaml:"icon"`
	Description  string       `yaml:"description"`
	Repository   string       `yaml:"repository"`
	Website      string       `yaml:"website"`
	Maintainers  []Maintainer `yaml:"maintainers"`
	Tags         []string     `yaml:"tags,omitempty"`
	Dependencies any          `yaml:"dependencies,omitempty"`
	Capabilities []string     `yaml:"capabilities"`
	Theme        *Theme       `yaml:"theme,omitempty"`
}

type Maintainer struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type Theme struct {
	Colors map[string]string `yaml:"colors,omitempty"`
}

// LoadPlugin loads and parses plugin.yaml, returning structured metadata
func LoadPluginMetadata(path string) (*PluginMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin metadata: %w", err)
	}

	var meta PluginMetadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.yaml: %w", err)
	}

	return &meta, nil
}

// Validate checks for required fields
func (m *PluginMetadata) Validate() error {
	var missing []string

	if m.ID == "" {
		missing = append(missing, "id")
	}
	if m.Name == "" {
		missing = append(missing, "name")
	}
	if m.Version == "" {
		missing = append(missing, "version")
	}
	if m.Description == "" {
		missing = append(missing, "description")
	}
	if m.Repository == "" {
		missing = append(missing, "repository")
	}
	if m.Website == "" {
		missing = append(missing, "website")
	}
	if len(m.Maintainers) == 0 {
		missing = append(missing, "maintainers")
	}
	if len(m.Capabilities) == 0 {
		missing = append(missing, "capabilities")
	}

	if len(missing) > 0 {
		return fmt.Errorf("plugin.yaml is missing required fields: %v", missing)
	}
	return nil
}

// SetVersion sets the version and returns updated YAML
func (m *PluginMetadata) SetVersion(version string) {
	m.Version = version
}

// Save writes the plugin.yaml back out to disk (optional step)
func (m *PluginMetadata) Save(path string) error {
	out, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal plugin.yaml: %w", err)
	}
	return os.WriteFile(path, out, 0644)
}
