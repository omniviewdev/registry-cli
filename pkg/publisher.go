package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/omniviewdev/registry-cli/pkg/types"
)

// Publisher is responsible for publishing a new version of a plugin to a registry. Currently,
// registries must be an aws S3 object store.
type Publisher struct {
	ctx      context.Context
	s3Client *s3.Client
	bucket   string
}

type PublisherOpts struct {
	Bucket  string
	Version string
}

func (p *PublisherOpts) Defaulter() {
	if p == nil {
		p = &PublisherOpts{}
	}

	if p.Bucket == "" {
		p.Bucket = os.Getenv("AWS_S3_BUCKET")
	}
}

// NewPublisher published a new release to the registry
func NewPublisher(ctx context.Context, opts PublisherOpts) (*Publisher, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.New(
			"couldn't load default configuration, have you set up your AWS account?",
		)
	}
	s3Client := s3.NewFromConfig(sdkConfig)

	opts.Defaulter()

	return &Publisher{
		ctx:      ctx,
		s3Client: s3Client,
		bucket:   opts.Bucket,
	}, nil
}

// Publish runs a publish of the plugin with the opts given. Used for publishing a version
// with all builds of the plugin in one command.
func (p *Publisher) Publish(ctx context.Context, opts types.PublishOpts) error {
	releases := opts.ToReleases()
	for _, release := range releases {
		releasePath, err := p.Upload(ctx, release)
		if err != nil {
			return err
		}

		fmt.Printf("uploaded release %s: %s\n", release, releasePath)
	}

	return nil
}

// Upload uploads the release to the location given the opts
func (p *Publisher) Upload(
	ctx context.Context,
	release types.Release,
) (string, error) {
	file, err := os.Open(release.Path)
	if err != nil {
		return "", fmt.Errorf("couldn't open file %v to upload: %v", release.Path, err)
	}

	fmt.Printf("uploading release to %s...\n", release.BucketPath())

	defer file.Close()
	_, err = p.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(release.BucketPath()),
		Body:   file,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			return "", fmt.Errorf(
				"error while uploading object to %s: the object is too large",
				p.bucket,
			)
		}

		return "", fmt.Errorf(
			"couldn't upload file %v to %v:%v: %v",
			release.Path,
			p.bucket,
			release.BucketPath(),
			err,
		)
	}
	err = s3.NewObjectExistsWaiter(p.s3Client).Wait(
		ctx, &s3.HeadObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(release.BucketPath())}, time.Minute)
	if err != nil {
		return "", fmt.Errorf("failed attempt to wait for object %s to exist", release.BucketPath())
	}

	return release.BucketPath(), nil
}
