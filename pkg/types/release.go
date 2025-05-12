package types

import "fmt"

type Release struct {
	Plugin  string
	Version string
	OS      string
	Arch    string
	Path    string
}

// Returns the path in the bucket to the release
func (r Release) BucketPath() string {
	return fmt.Sprintf("%s/%s/%s-%s.tar.gz", r.Plugin, r.Version, r.OS, r.Arch)
}

// Returns the architecture key used for the index (amongst other things)
func (r Release) OSArch() string {
	return fmt.Sprintf("%s_%s", r.OS, r.Arch)
}

func (r Release) String() string {
	return fmt.Sprintf("%s [%s/%s] - %s", r.Plugin, r.OS, r.Arch, r.Version)
}

type PublishOpts struct {
	// The plugin we're updating
	Plugin string

	// The version we're indexing
	Version string

	// Metadata stores the path to the metadata file
	MetadataPath string

	// Path to a darwin/arm64 build
	DarwinARM64 string

	// Path to a darwin/amd64 build
	DarwinAMD64 string

	// Path to a windows/arm64 build
	WindowsARM64 string

	// Path to a windows/amd64 build
	WindowsAMD64 string

	// Path to a linux/arm64 build
	LinuxARM64 string

	// Path to a linux/amd64 build
	LinuxAMD64 string
}

func (p PublishOpts) ToReleases() []Release {
	// build out our release objects
	releases := make([]Release, 0)

	if p.DarwinARM64 != "" {
		releases = append(releases, Release{
			Plugin:  p.Plugin,
			Version: p.Version,
			OS:      "darwin",
			Arch:    "arm64",
			Path:    p.DarwinARM64,
		})
	}
	if p.DarwinAMD64 != "" {
		releases = append(releases, Release{
			Plugin:  p.Plugin,
			Version: p.Version,
			OS:      "darwin",
			Arch:    "amd64",
			Path:    p.DarwinAMD64,
		})
	}
	if p.WindowsARM64 != "" {
		releases = append(releases, Release{
			Plugin:  p.Plugin,
			Version: p.Version,
			OS:      "windows",
			Arch:    "arm64",
			Path:    p.WindowsARM64,
		})
	}
	if p.WindowsAMD64 != "" {
		releases = append(releases, Release{
			Plugin:  p.Plugin,
			Version: p.Version,
			OS:      "windows",
			Arch:    "amd64",
			Path:    p.WindowsAMD64,
		})
	}
	if p.LinuxARM64 != "" {
		releases = append(releases, Release{
			Plugin:  p.Plugin,
			Version: p.Version,
			OS:      "linux",
			Arch:    "arm64",
			Path:    p.LinuxARM64,
		})
	}
	if p.LinuxAMD64 != "" {
		releases = append(releases, Release{
			Plugin:  p.Plugin,
			Version: p.Version,
			OS:      "linux",
			Arch:    "amd64",
			Path:    p.LinuxAMD64,
		})
	}

	return releases
}
