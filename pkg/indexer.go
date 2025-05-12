package pkg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/omniviewdev/registry-cli/pkg/types"
)

// Indexer is responsible for updating the index based on a release
type Indexer struct {
	ctx      context.Context
	s3Client *s3.Client
	bucket   string
}

type IndexerOpts struct {
	Bucket string
}

func (p *IndexerOpts) Defaulter() {
	if p == nil {
		p = &IndexerOpts{}
	}

	if p.Bucket == "" {
		p.Bucket = os.Getenv("AWS_S3_BUCKET")
	}
}

// NewIndexer creates a new indexing service for updating after a release
func NewIndexer(ctx context.Context, opts IndexerOpts) (*Indexer, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.New(
			"couldn't load default configuration, have you set up your AWS account?",
		)
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	opts.Defaulter()

	return &Indexer{
		ctx:      ctx,
		s3Client: s3Client,
		bucket:   opts.Bucket,
	}, nil
}

// UpdateIndex updates the plugin index with the new release
func (i *Indexer) UpdateIndex(ctx context.Context, opts types.PublishOpts) error {
	// get the metadata file
	metadata := types.LoadMetadata(opts.MetadataPath)
	index, err := i.getPluginIndex(ctx, opts.Plugin)
	if err != nil {
		return err
	}

	// build out our release objects
	releases := opts.ToReleases()
	pluginIndex := i.updateIndex(index, releases, metadata)
	_, err = i.setPluginIndex(ctx, pluginIndex)
	if err != nil {
		return err
	}

	// update the registry index
	registryIndex, err := i.getRegistryIndex(ctx)
	if err != nil {
		return err
	}

	found := false
	for idx, plugin := range registryIndex.Plugins {
		if plugin.ID == pluginIndex.ID {
			found = true

			registryIndex.Plugins[idx] = types.RegistryIndexPlugins{
				ID:            pluginIndex.ID,
				Name:          pluginIndex.Name,
				Icon:          pluginIndex.Icon,
				Description:   pluginIndex.Description,
				Official:      true,
				LatestVersion: pluginIndex.LatestVersion,
			}

			break
		}
	}

	if !found {
		registryIndex.Plugins = append(registryIndex.Plugins, types.RegistryIndexPlugins{
			ID:            pluginIndex.ID,
			Name:          pluginIndex.Name,
			Icon:          pluginIndex.Icon,
			Description:   pluginIndex.Description,
			Official:      true,
			LatestVersion: pluginIndex.LatestVersion,
		})
	}

	_, err = i.setRegistryIndex(ctx, registryIndex)
	if err != nil {
		return err
	}

	// all good!
	return nil
}

// updateIndex updates the index based on the plugin and passed in versions. It is expected the
// releases are all the same version and of different architectures.
func (i *Indexer) updateIndex(
	index types.PluginIndex,
	releases []types.Release,
	metadata types.PluginMeta,
) types.PluginIndex {
	if len(releases) < 1 {
		panic("cannot submit an empty number of releases")
	}

	versionInfo := types.PluginVersionInformation{
		Version:       releases[0].Version,
		Architectures: make(map[string]types.PluginArchitectureInformation, len(releases)),
		Created:       time.Now(),
		Updated:       time.Now(),
		Metadata:      metadata,
	}

	// build the versions out
	for _, release := range releases {
		if release.Plugin != index.ID {
			// not sure how we got here, but don't let this keep going
			log.Printf("got release that wasn't part of plugin '%s'\n", release.Plugin)
			continue
		}
		info := types.PluginArchitectureInformation{
			Checksum:    "TODO",
			DownloadURL: release.BucketPath(),
		}

		// Calculate Checksum
		f, err := os.Open(release.Path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}
		info.Checksum = hex.EncodeToString(h.Sum(nil))

		// Calculate file info
		fileInfo, err := os.Stat(release.Path)
		if err != nil {
			fmt.Println("Failed to calculate size: ", err)
		} else {
			info.Size = fileInfo.Size()
		}

		versionInfo.Architectures[release.OSArch()] = info
	}

	index.LatestVersion = versionInfo
	index.Versions = append(index.Versions, versionInfo)

	// update the info using the metadata
	index.Description = metadata.Description
	index.Icon = metadata.Icon
	index.Name = metadata.Name

	return index
}

// getPluginIndex returns a plugin index either from the bucket if it exists, or a new one
func (i *Indexer) getPluginIndex(ctx context.Context, plugin string) (types.PluginIndex, error) {
	// first check the s3 bucket
	result, err := i.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(i.bucket),
		Key:    aws.String(fmt.Sprintf("%s/index.json", plugin)),
	})
	if err != nil {
		var noKey *s3types.NoSuchKey
		if !errors.As(err, &noKey) {
			return types.PluginIndex{}, fmt.Errorf("couldn't get plugin index: %v", err)
		}

		// don't have an index yet, create one and return it (though it will be minimal)
		return types.PluginIndex{
			RegistryIndexPlugins: types.RegistryIndexPlugins{
				ID:   plugin,
				Name: plugin,
			},
		}, nil
	}

	// at this point we have an index

	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return types.PluginIndex{}, fmt.Errorf("couldn't read object body: %v", err)
	}

	var index types.PluginIndex
	if err := json.Unmarshal(body, &index); err != nil {
		return index, fmt.Errorf("couldn't decode object body to json: %v", err)
	}

	return index, nil
}

// getRegistryIindex returns the registry index
func (i *Indexer) getRegistryIndex(ctx context.Context) (types.RegistryIndex, error) {
	// first check the s3 bucket
	result, err := i.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(i.bucket),
		Key:    aws.String("index.json"),
	})
	if err != nil {
		var noKey *s3types.NoSuchKey
		if !errors.As(err, &noKey) {
			return types.RegistryIndex{}, fmt.Errorf("couldn't get registry index: %v", err)
		}

		// don't have an index yet, create one and return it (though it will be minimal)
		return types.RegistryIndex{
			Plugins: make([]types.RegistryIndexPlugins, 0),
		}, nil
	}

	// at this point we have an index

	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return types.RegistryIndex{}, fmt.Errorf("couldn't read object body: %v", err)
	}

	var index types.RegistryIndex
	if err := json.Unmarshal(body, &index); err != nil {
		return index, fmt.Errorf("couldn't decode object body to json: %v", err)
	}

	return index, nil
}

// setPluginIndex updates the plugin index within the storage bucket
func (i *Indexer) setPluginIndex(ctx context.Context, index types.PluginIndex) (string, error) {
	b, err := json.Marshal(index)
	if err != nil {
		return "", fmt.Errorf("failed to upload plugin index: %v", err)
	}

	fmt.Printf("uploading plugin index to %s...\n", index.BucketPath())
	return i.store(ctx, b, index.BucketPath())
}

// setGlobalIndex updates the global index within the storage bucket
func (i *Indexer) setRegistryIndex(ctx context.Context, index types.RegistryIndex) (string, error) {
	b, err := json.Marshal(index)
	if err != nil {
		return "", fmt.Errorf("failed to upload plugin index: %v", err)
	}

	fmt.Printf("uploading registry index...\n")
	return i.store(ctx, b, "index.json")
}

// store stores into the S3 bucket
func (i *Indexer) store(ctx context.Context, b []byte, bucketPath string) (string, error) {
	_, err := i.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(i.bucket),
		Key:    aws.String(bucketPath),
		Body:   bytes.NewBuffer(b),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			return "", fmt.Errorf(
				"error while uploading object to %s: the object is too large",
				i.bucket,
			)
		}

		return "", fmt.Errorf(
			"couldn't upload plugin index to %v:%v: %v",
			i.bucket,
			bucketPath,
			err,
		)
	}
	err = s3.NewObjectExistsWaiter(i.s3Client).Wait(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(i.bucket),
			Key:    aws.String(bucketPath),
		}, time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed attempt to wait for object %s to exist", bucketPath)
	}

	return bucketPath, nil
}
