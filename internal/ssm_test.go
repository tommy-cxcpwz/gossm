package internal

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindInstances_ValidConfig_ReturnsNoError(t *testing.T) {
	cfg, err := NewSharedConfig(context.Background(), mockProfile,
		[]string{config.DefaultSharedConfigFilename()},
		[]string{config.DefaultSharedCredentialsFilename()})
	require.NoError(t, err)

	tests := []struct {
		name  string
		ctx   context.Context
		cfg   aws.Config
		isErr bool
	}{
		{
			name:  "Success_ReturnsInstances",
			ctx:   context.Background(),
			cfg:   cfg,
			isErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindInstances(tt.ctx, tt.cfg)

			assert.Equal(t, tt.isErr, err != nil)
		})
	}
}

func TestFindInstanceIdsWithConnectedSSM_ValidConfig_ReturnsNoError(t *testing.T) {
	cfg, err := NewSharedConfig(context.Background(), mockProfile,
		[]string{config.DefaultSharedConfigFilename()},
		[]string{config.DefaultSharedCredentialsFilename()})
	require.NoError(t, err)

	tests := []struct {
		name  string
		ctx   context.Context
		cfg   aws.Config
		isErr bool
	}{
		{
			name:  "Success_ReturnsInstanceIds",
			ctx:   context.Background(),
			cfg:   cfg,
			isErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FindInstanceIdsWithConnectedSSM(tt.ctx, tt.cfg)

			assert.Equal(t, tt.isErr, err != nil)
		})
	}
}

func TestPrintReady_Called_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintReady("test-cmd", "us-east-1", "i-1234567890abcdef0")
	})
}
