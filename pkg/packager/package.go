package packager

import (
	"fmt"
	"os"
	"path/filepath"
)

type PackOpts struct {
	PluginDir string
	Version   string
	OutDir    string
	Clean     bool
}

// RunPackCommand runs the packaging step
func RunPackCommand(opts PackOpts) (*PluginMetadata, error) {
	if opts.OutDir == "" {
		return nil, fmt.Errorf("cannot build to empty directory")
	}
	if opts.OutDir == "/" {
		return nil, fmt.Errorf("DANGER: You supplied the root directory as the output directory")
	}

	if opts.Clean {
		if err := os.RemoveAll(opts.OutDir); err != nil {
			return nil, fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	meta, err := LoadPluginMetadata(filepath.Join(opts.PluginDir, "plugin.yaml"))
	if err != nil {
		return nil, fmt.Errorf("invalid plugin.yaml: %w", err)
	}

	if err := meta.Validate(); err != nil {
		return nil, err
	}

	meta.SetVersion(opts.Version)

	// You can optionally write it back out before packaging
	if err := meta.Save(filepath.Join(opts.PluginDir, "plugin.yaml")); err != nil {
		return nil, err
	}

	// Supported platforms
	targets := []Platform{
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"windows", "amd64"},
		{"windows", "arm64"},
	}

	// Run all builds concurrently
	buildResults := BuildAll(opts.PluginDir, opts.Version, opts.OutDir, targets)

	// Compress each successful build
	for _, result := range buildResults {
		if result.Err != nil {
			fmt.Printf("❌ Build failed for %s: %v\n", result.Platform, result.Err)
			continue
		}
		out := filepath.Join(
			opts.PluginDir,
			fmt.Sprintf("%s/%s.tar.gz", opts.OutDir, result.Platform.Key()),
		)
		if _, _, err := TarGz(result.OutputDir, out); err != nil {
			return nil, fmt.Errorf("compression failed for %s: %w", result.Platform.Key(), err)
		}
		fmt.Printf("✅ Packaged %s → %s\n", result.Platform.Key(), out)
	}

	fmt.Printf("\nSuccessfully packaged plugin for distribution\n")

	return meta, nil
}
