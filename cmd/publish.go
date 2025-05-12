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

	"github.com/omniviewdev/registry-cli/pkg"
	"github.com/omniviewdev/registry-cli/pkg/types"
	"github.com/spf13/cobra"
)

var (
	bucket        string
	metadata      string
	darwin_arm64  string
	darwin_amd64  string
	windows_arm64 string
	windows_amd64 string
	linux_arm64   string
	linux_amd64   string
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish [plugin] [version]",
	Short: "Publish a new version of your plugin",
	Long: `Push a new version of a plugin to the registry. This action updates
the indexes within the registry to show the new version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch len(args) {
		case 0:
			return cmd.Help()
		case 1:
			// TODO: validate the version string
			return fmt.Errorf(
				"Missing version string. Please provide as the second argument to 'publish'",
			)
		}

		opts := types.PublishOpts{
			Plugin:       args[0],
			Version:      args[1],
			MetadataPath: metadata,
			DarwinAMD64:  darwin_amd64,
			DarwinARM64:  darwin_arm64,
			WindowsAMD64: windows_amd64,
			WindowsARM64: windows_arm64,
			LinuxAMD64:   linux_amd64,
			LinuxARM64:   linux_arm64,
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

		if err := publisher.Publish(cmd.Context(), opts); err != nil {
			return err
		}
		if err := indexer.UpdateIndex(cmd.Context(), opts); err != nil {
			return err
		}

		fmt.Printf("published new version: %v\n", opts)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	publishCmd.Flags().StringVarP(&bucket, "bucket", "b", "", "bucket to upload to")
	publishCmd.Flags().StringVarP(&metadata, "metadata", "m", "", "path to plugin metadata file")
	publishCmd.Flags().StringVar(&darwin_arm64, "darwin_arm64", "", "path to a darwin/arm64 build")
	publishCmd.Flags().StringVar(&darwin_amd64, "darwin_amd64", "", "path to a darwin/amd64 build")
	publishCmd.Flags().
		StringVar(&windows_arm64, "windows_arm64", "", "path to a windows/arm64 build")
	publishCmd.Flags().
		StringVar(&windows_amd64, "windows_amd64", "", "path to a windows/amd64 build")
	publishCmd.Flags().StringVar(&linux_arm64, "linux_arm64", "", "path to a linux/arm64 build")
	publishCmd.Flags().StringVar(&linux_amd64, "linux_amd64", "", "path to a linux/amd64 build")
}
