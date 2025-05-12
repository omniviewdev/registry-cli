package types

// RegistryIndex is the file at the root of the plugin registry that exposes information about
// what plugins are available, for what architectures, and what versions.
type RegistryIndex struct {
	// Plugins lists the plugins available along with their metadata for viewing within omniview
	Plugins []RegistryIndexPlugins `json:"plugins"`
}

// RegistryIndexPlugins
type RegistryIndexPlugins struct {
	ID            string                   `json:"id"`
	Name          string                   `json:"name"`
	Icon          string                   `json:"icon"`
	Description   string                   `json:"description"`
	Official      bool                     `json:"official"`
	LatestVersion PluginVersionInformation `json:"latest_version"`
}
