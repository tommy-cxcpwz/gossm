package internal

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSharedConfig_ValidProfile_ReturnsNoError(t *testing.T) {
	_, err := NewSharedConfig(context.Background(), mockProfile,
		[]string{config.DefaultSharedConfigFilename()},
		[]string{config.DefaultSharedCredentialsFilename()})

	assert.NoError(t, err)
}

func TestNewSharedConfig_NilContext_ReturnsError(t *testing.T) {
	//nolint:staticcheck // testing nil context handling intentionally
	_, err := NewSharedConfig(nil, "default", nil, nil)

	assert.Error(t, err)
}

func TestNewSharedConfig_NonexistentProfile_ReturnsError(t *testing.T) {
	_, err := NewSharedConfig(context.Background(), "nonexistent-profile", []string{}, []string{})

	assert.Error(t, err)
}
