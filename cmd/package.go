/*
Copyright © 2025 Joshua Pare <jpare@omniview.dev>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/omniviewdev/registry-cli/pkg"
	"github.com/omniviewdev/registry-cli/pkg/packager"
	"github.com/omniviewdev/registry-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	clean   bool
	outdir  string
	version string
	publish bool
)

// packageCmd represents the package command
var packageCmd = &cobra.Command{
	Use:   "package [path]",
	Short: "Package a plugin for distribution",
	Long: `Package compiles the necessary binaries and files into the proper
location for uploading to the Omniview Plugin Registry.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch len(args) {
		case 0:
			// TODO: validate the version string
			return fmt.Errorf(
				"Missing path to plugin. Please provide as the first argument to 'package'",
			)
		}

		// if we're publishing too, make sure we've supplied
		if publish && bucket == "" {
			return fmt.Errorf("Must supply a bucket when --publish is set to true")
		}

		opts := packager.PackOpts{
			PluginDir: args[0],
			OutDir:    outdir,
			Version:   version,
			Clean:     clean,
		}

		meta, err := packager.RunPackCommand(opts)
		if err != nil {
			return err
		}

		if !publish {
			return nil
		}

		fmt.Println("Publishing to registry...")

		// we're going to also publish to the registry
		publishOpts := types.PublishOpts{
			Plugin:       meta.ID,
			Version:      meta.Version,
			MetadataPath: filepath.Join(args[0], "plugin.yaml"),
			DarwinAMD64:  filepath.Join(outdir, "darwin_amd64.tar.gz"),
			DarwinARM64:  filepath.Join(outdir, "darwin_arm64.tar.gz"),
			WindowsAMD64: filepath.Join(outdir, "windows_amd64.tar.gz"),
			WindowsARM64: filepath.Join(outdir, "windows_arm64.tar.gz"),
			LinuxAMD64:   filepath.Join(outdir, "linux_amd64.tar.gz"),
			LinuxARM64:   filepath.Join(outdir, "linux_arm64.tar.gz"),
		}

		indexer, err := pkg.NewIndexer(cmd.Context(), pkg.IndexerOpts{
			Bucket: bucket,
		})
		if err != nil {
			return err
		}

		publisher, err := pkg.NewPublisher(cmd.Context(), pkg.PublisherOpts{
			Bucket: bucket,
		})
		if err != nil {
			return err
		}

		if err := publisher.Publish(cmd.Context(), publishOpts); err != nil {
			return err
		}
		if err := indexer.UpdateIndex(cmd.Context(), publishOpts); err != nil {
			return err
		}

		fmt.Printf(
			"Published new plugin version: %s[%s]\n",
			publishOpts.Plugin,
			publishOpts.Version,
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(packageCmd)

	packageCmd.Flags().
		BoolVarP(&clean, "clean", "c", true, "Clean the output directory before packaging")
	packageCmd.Flags().
		StringVarP(&outdir, "out", "o", "build", "Output directory for the plugin packages")
	packageCmd.Flags().
		StringVarP(&version, "version", "v", "", "Version to use for the build. Defaults to what is in the plugin.yaml")

	packageCmd.Flags().
		BoolVarP(&publish, "publish", "p", false, "Publish the builds to the registry after building")
	packageCmd.Flags().
		StringVarP(&bucket, "bucket", "b", "", "Bucket to use when running with the 'publish' flag")
}
