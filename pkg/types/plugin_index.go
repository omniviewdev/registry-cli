package types

import (
	"fmt"
	"time"
)

// PluginIndex is the file at the root of the plugin folder that exposes information about
// what versions are available for a specific plugin and what architectures are supported.
type PluginIndex struct {
	RegistryIndexPlugins

	// Versions is the list of version available
	Versions []PluginVersionInformation `json:"versions"`
}

// BucketPath get's the bucket path for where the index should be located
func (i PluginIndex) BucketPath() string {
	return fmt.Sprintf("%s/index.json", i.ID)
}

type PluginVersionInformation struct {
	// Metadata is the metadata for this version
	Metadata PluginMeta `json:"metadata"`

	// Version is the semver string for the version provided
	Version string `json:"version"`

	// Stores links to the tarball for each architecture build
	Architectures map[string]PluginArchitectureInformation `json:"architectures"`

	// Created
	Created time.Time `json:"created"`

	// Updated
	Updated time.Time `json:"updated"`
}

type PluginArchitectureInformation struct {
	// Checksum is the checksum to expect for the plugin
	Checksum string `json:"checksum"`

	// DownloadURL is the url for which to download the tarball
	DownloadURL string `json:"download_url"`

	// Size is the calculated size of the tarball in bytes
	Size int64 `json:"size"`
}
