package types

import (
	"io"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

type PluginMetaFormat int

const (
	// PluginMetaFormatYAML is the YAML format for a plugin config.
	PluginMetaFormatYAML PluginMetaFormat = iota
	// PluginMetaFormatJSON is the JSON format for a plugin config.
	PluginMetaFormatJSON
)

// PluginMeta is the plugin description file located at the root of a plugin.
type PluginMeta struct {
	ID           string             `json:"id"           yaml:"id"`
	Version      string             `json:"version"      yaml:"version"`
	Name         string             `json:"name"         yaml:"name"`
	Icon         string             `json:"icon"         yaml:"icon"`
	Description  string             `json:"description"  yaml:"description"`
	Repository   string             `json:"repository"   yaml:"repository"`
	Website      string             `json:"website"      yaml:"website"`
	Markdown     string             `json:"-"            yaml:"-"`
	Maintainers  []PluginMaintainer `json:"maintainers"  yaml:"maintainers"`
	Tags         []string           `json:"tags"         yaml:"tags"`
	Dependencies []string           `json:"dependencies" yaml:"dependencies"`
	Capabilities []string           `json:"capabilities" yaml:"capabilities"`
	Theme        PluginTheme        `json:"theme"        yaml:"theme"`
}

// HasUICapabilities checks if the plugin has UI capabilities. This is used
// to verify plugin loading and staring.
func (m *PluginMeta) HasUICapabilities() bool {
	for _, capability := range m.Capabilities {
		if capability == "ui" {
			return true
		}
	}
	return false
}

// HasBackendCapabilities checks if the plugin has UI capabilities. This is used
// to verify plugin loading and staring.
func (m *PluginMeta) HasBackendCapabilities() bool {
	caps := []string{"resource", "exec", "networker", "settings"}

	for _, capability := range m.Capabilities {
		if slices.Contains(caps, capability) {
			return true
		}
	}
	return false
}

type PluginMaintainer struct {
	Name  string `json:"name"  yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

type PluginTheme struct {
	Colors struct {
		Primary   string `json:"primary"   yaml:"primary"`
		Secondary string `json:"secondary" yaml:"secondary"`
		Tertiary  string `json:"tertiary"  yaml:"tertiary"`
	} `json:"colors" yaml:"colors"`
}

type PluginComponents struct {
	Resource []PluginResourceComponent `json:"resource" yaml:"resource"`
}

type PluginComponentArea string

const (
	PluginComponentAreaEditor  PluginComponentArea = "EDITOR"
	PluginComponentAreaSidebar PluginComponentArea = "SIDEBAR"
)

type PluginResourceComponent struct {
	Name           string              `json:"name"      yaml:"name"`
	Plugin         string              `json:"plugin"    yaml:"plugin"`
	Area           PluginComponentArea `json:"area"      yaml:"area"`
	Resources      []string            `json:"resources" yaml:"resources"`
	ExtensionPoint string              `json:"extension" yaml:"extension"`
}

func (c *PluginMeta) Load(reader io.Reader) error {
	return yaml.NewDecoder(reader).Decode(c)
}

// LoadMetadata
func LoadMetadata(path string) PluginMeta {
	file, err := os.Open(path)
	if err != nil {
		return PluginMeta{}
	}

	meta := PluginMeta{}
	if err := yaml.NewDecoder(file).Decode(&meta); err != nil {
		return PluginMeta{}
	}

	return meta
}

// LoadMarkdown loads a plugin Markdown file from a given path (if it exists)
// into the plugin config.
func (c *PluginMeta) LoadMarkdown(path string) error {
	// load markdown file from path and set c.Markdown to the loaded markdown
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// read markdown file
	markdown, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	c.Markdown = string(markdown)
	return nil
}
