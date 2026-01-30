package internal

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindInstances_ValidConfig_ReturnsNoError(t *testing.T) {
	cfg, err := NewSharedConfig(context.Background(), mockProfile,
		[]string{config.DefaultSharedConfigFilename()},
		[]string{config.DefaultSharedCredentialsFilename()})
	require.NoError(t, err)

	ssmClient := ssm.NewFromConfig(cfg)
	ec2Client := ec2.NewFromConfig(cfg)

	_, err = FindInstances(context.Background(), ssmClient, ec2Client)

	assert.NoError(t, err)
}

func TestFindInstanceIdsWithConnectedSSM_ValidConfig_ReturnsNoError(t *testing.T) {
	cfg, err := NewSharedConfig(context.Background(), mockProfile,
		[]string{config.DefaultSharedConfigFilename()},
		[]string{config.DefaultSharedCredentialsFilename()})
	require.NoError(t, err)

	client := ssm.NewFromConfig(cfg)

	_, err = FindInstanceIdsWithConnectedSSM(context.Background(), client)

	assert.NoError(t, err)
}

func TestPrintReady_Called_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		PrintReady("test-cmd", "us-east-1", "i-1234567890abcdef0")
	})
}
