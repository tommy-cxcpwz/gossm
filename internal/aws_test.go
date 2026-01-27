package internal

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		ctx     context.Context
		key     string
		secret  string
		token   string
		region  string
		roleArn string
		isErr   bool
	}{
		"fail":    {isErr: true},
		"success": {ctx: context.Background(), key: mockAwsKey, secret: mockAwsSecret, region: mockRegion, isErr: false},
	}

	for _, t := range tests {
		_, err := NewConfig(t.ctx, t.key, t.secret, t.token, t.region, t.roleArn)
		assert.Equal(t.isErr, err != nil)
	}
}

func TestNewSharedConfig(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		ctx               context.Context
		profile           string
		sharedCredentials []string
		sharedConfigs     []string
		isErr             bool
	}{
		"fail": {isErr: true},
		"success": {
			ctx:               context.Background(),
			profile:           mockProfile,
			sharedConfigs:     []string{config.DefaultSharedConfigFilename()},
			sharedCredentials: []string{config.DefaultSharedCredentialsFilename()},
			isErr:             false},
	}

	for _, t := range tests {
		_, err := NewSharedConfig(t.ctx, t.profile, t.sharedConfigs, t.sharedCredentials)
		assert.Equal(t.isErr, err != nil)
	}
}

func TestNewConfigWithRegion(t *testing.T) {
	assert := assert.New(t)

	cfg, err := NewConfig(context.Background(), mockAwsKey, mockAwsSecret, "", "us-west-2", "")
	assert.NoError(err)
	assert.Equal("us-west-2", cfg.Region)
}

func TestNewConfigWithToken(t *testing.T) {
	assert := assert.New(t)

	cfg, err := NewConfig(context.Background(), mockAwsKey, mockAwsSecret, "test-token", mockRegion, "")
	assert.NoError(err)
	assert.NotNil(cfg)
}

func TestNewConfigWithRoleArn(t *testing.T) {
	assert := assert.New(t)

	cfg, err := NewConfig(context.Background(), mockAwsKey, mockAwsSecret, "", mockRegion, "arn:aws:iam::123456789012:role/TestRole")
	assert.NoError(err)
	assert.NotNil(cfg)
	assert.NotNil(cfg.Credentials)
}

func TestNewConfigNilContext(t *testing.T) {
	assert := assert.New(t)

	//nolint:staticcheck // testing nil context handling intentionally
	_, err := NewConfig(nil, "key", "secret", "", "us-east-1", "")
	assert.Error(err)
}

func TestNewSharedConfigNilContext(t *testing.T) {
	assert := assert.New(t)

	//nolint:staticcheck // testing nil context handling intentionally
	_, err := NewSharedConfig(nil, "default", nil, nil)
	assert.Error(err)
}

func TestNewConfigDefaultCredentials(t *testing.T) {
	assert := assert.New(t)

	// Test loading default credentials (empty key/secret)
	cfg, err := NewConfig(context.Background(), "", "", "", mockRegion, "")
	assert.NoError(err)
	assert.NotNil(cfg)
}

func TestNewSharedConfigEmptyFiles(t *testing.T) {
	assert := assert.New(t)

	// Test with empty file lists - should fail as no credentials are available
	_, err := NewSharedConfig(context.Background(), "nonexistent-profile", []string{}, []string{})
	assert.Error(err)
}

func TestNewConfigWithAllParams(t *testing.T) {
	assert := assert.New(t)

	// Test with all parameters filled
	cfg, err := NewConfig(context.Background(), mockAwsKey, mockAwsSecret, "token123", "eu-west-1", "")
	assert.NoError(err)
	assert.Equal("eu-west-1", cfg.Region)
}

func TestNewConfigNoRegion(t *testing.T) {
	assert := assert.New(t)

	// Test with no region - should use default
	cfg, err := NewConfig(context.Background(), mockAwsKey, mockAwsSecret, "", "", "")
	assert.NoError(err)
	assert.NotNil(cfg)
}
