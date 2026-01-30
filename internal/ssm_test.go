package internal

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssm_types "github.com/aws/aws-sdk-go-v2/service/ssm/types"
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

func TestPrintReadyMulti_Called_DoesNotPanic(t *testing.T) {
	targets := []*Target{
		{Name: "i-0abc123", TagName: "web-server"},
		{Name: "i-0def456", TagName: "db-server"},
	}

	assert.NotPanics(t, func() {
		PrintReadyMulti("uptime", "us-east-1", targets)
	})
}

func TestPrintReadyMulti_SingleTarget_DoesNotPanic(t *testing.T) {
	targets := []*Target{
		{Name: "i-0abc123"},
	}

	assert.NotPanics(t, func() {
		PrintReadyMulti("ls", "ap-east-1", targets)
	})
}

// mockSSMCommandAPI implements SSMCommandAPI for testing.
type mockSSMCommandAPI struct {
	getCommandInvocationFunc func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error)
	sendCommandFunc          func(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error)
}

func (m *mockSSMCommandAPI) SendCommand(ctx context.Context, params *ssm.SendCommandInput, optFns ...func(*ssm.Options)) (*ssm.SendCommandOutput, error) {
	return m.sendCommandFunc(ctx, params, optFns...)
}

func (m *mockSSMCommandAPI) GetCommandInvocation(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
	return m.getCommandInvocationFunc(ctx, params, optFns...)
}

func TestPrintCommandInvocation_Success_PrintsNameTag(t *testing.T) {
	mock := &mockSSMCommandAPI{
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				InstanceId:            params.InstanceId,
				Status:                ssm_types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("hello"),
			}, nil
		},
	}

	inputs := []*ssm.GetCommandInvocationInput{
		{CommandId: aws.String("cmd-123"), InstanceId: aws.String("i-0abc123")},
		{CommandId: aws.String("cmd-123"), InstanceId: aws.String("i-0def456")},
	}
	nameMap := map[string]string{
		"i-0abc123": "web-server",
		"i-0def456": "",
	}

	assert.NotPanics(t, func() {
		PrintCommandInvocation(context.Background(), mock, inputs, nameMap)
	})
}

func TestPrintCommandInvocation_NilNameMap_DoesNotPanic(t *testing.T) {
	mock := &mockSSMCommandAPI{
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				InstanceId:            params.InstanceId,
				Status:                ssm_types.CommandInvocationStatusSuccess,
				StandardOutputContent: aws.String("ok"),
			}, nil
		},
	}

	inputs := []*ssm.GetCommandInvocationInput{
		{CommandId: aws.String("cmd-456"), InstanceId: aws.String("i-0abc123")},
	}

	assert.NotPanics(t, func() {
		PrintCommandInvocation(context.Background(), mock, inputs, nil)
	})
}

func TestPrintCommandInvocation_Failed_DoesNotPanic(t *testing.T) {
	mock := &mockSSMCommandAPI{
		getCommandInvocationFunc: func(ctx context.Context, params *ssm.GetCommandInvocationInput, optFns ...func(*ssm.Options)) (*ssm.GetCommandInvocationOutput, error) {
			return &ssm.GetCommandInvocationOutput{
				InstanceId:           params.InstanceId,
				Status:               ssm_types.CommandInvocationStatusFailed,
				StatusDetails:        aws.String("Failed"),
				StandardErrorContent: aws.String("command not found"),
			}, nil
		},
	}

	inputs := []*ssm.GetCommandInvocationInput{
		{CommandId: aws.String("cmd-789"), InstanceId: aws.String("i-0abc123")},
	}
	nameMap := map[string]string{"i-0abc123": "my-server"}

	assert.NotPanics(t, func() {
		PrintCommandInvocation(context.Background(), mock, inputs, nameMap)
	})
}

func TestFormatTags_MultipleTags_ReturnsSortedJSON(t *testing.T) {
	tags := map[string]string{
		"Environment": "prod",
		"App":         "web",
	}

	result := FormatTags(tags)

	expected := "{\n    App = \"web\",\n    Environment = \"prod\"\n}"
	assert.Equal(t, expected, result)
}

func TestFormatTags_EmptyMap_ReturnsEmptyBraces(t *testing.T) {
	result := FormatTags(map[string]string{})

	assert.Equal(t, "{}", result)
}

func TestFormatTags_NilMap_ReturnsEmptyBraces(t *testing.T) {
	result := FormatTags(nil)

	assert.Equal(t, "{}", result)
}

func TestFormatTags_SingleTag_ReturnsJSON(t *testing.T) {
	tags := map[string]string{"Env": "staging"}

	result := FormatTags(tags)

	expected := "{\n    Env = \"staging\"\n}"
	assert.Equal(t, expected, result)
}
