package internal

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// NewSharedConfig creates a config for accessing AWS that is based on shared files, such as credentials file.
func NewSharedConfig(ctx context.Context, profile string, sharedConfigFiles, sharedCredentialsFiles []string) (aws.Config, error) {
	if ctx == nil {
		return aws.Config{}, WrapError(ErrInvalidParams)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithSharedConfigFiles(sharedConfigFiles),
		config.WithSharedCredentialsFiles(sharedCredentialsFiles),
		// Disable EC2 IMDS to avoid 1-2 second timeout when not on EC2
		config.WithEC2IMDSClientEnableState(imds.ClientDisabled),
	)
	if err != nil {
		return aws.Config{}, WrapError(err)
	}

	return cfg, nil
}
