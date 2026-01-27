package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestFindInstances(t *testing.T) {
	assert := assert.New(t)

	cfg, err := NewConfig(context.Background(), "", "", "", "", "")
	assert.NoError(err)

	tests := map[string]struct {
		ctx   context.Context
		cfg   aws.Config
		isErr bool
	}{
		"success": {
			ctx:   context.Background(),
			cfg:   cfg,
			isErr: false,
		},
	}

	for _, t := range tests {
		result, err := FindInstances(t.ctx, t.cfg)
		assert.Equal(t.isErr, err != nil)
		fmt.Println(len(result))
	}
}
func TestFindInstanceIdsWithConnectedSSM(t *testing.T) {
	assert := assert.New(t)

	cfg, err := NewConfig(context.Background(), "", "", "", "", "")
	assert.NoError(err)

	tests := map[string]struct {
		ctx   context.Context
		cfg   aws.Config
		isErr bool
	}{
		"success": {
			ctx:   context.Background(),
			cfg:   cfg,
			isErr: false,
		},
	}

	for _, t := range tests {
		result, err := FindInstanceIdsWithConnectedSSM(t.ctx, t.cfg)
		assert.Equal(t.isErr, err != nil)
		fmt.Println(len(result))
	}
}

func TestTargetStruct(t *testing.T) {
	assert := assert.New(t)

	target := &Target{
		Name:          "i-1234567890abcdef0",
		PublicDomain:  "ec2-1-2-3-4.compute.amazonaws.com",
		PrivateDomain: "ip-10-0-0-1.ec2.internal",
	}

	assert.Equal("i-1234567890abcdef0", target.Name)
	assert.Equal("ec2-1-2-3-4.compute.amazonaws.com", target.PublicDomain)
	assert.Equal("ip-10-0-0-1.ec2.internal", target.PrivateDomain)
}

func TestUserStruct(t *testing.T) {
	assert := assert.New(t)

	user := &User{Name: "ubuntu"}
	assert.Equal("ubuntu", user.Name)
}

func TestRegionStruct(t *testing.T) {
	assert := assert.New(t)

	region := &Region{Name: "us-east-1"}
	assert.Equal("us-east-1", region.Name)
}

func TestDefaultAwsRegions(t *testing.T) {
	assert := assert.New(t)

	// Test that default regions include expected values
	assert.Contains(defaultAwsRegions, "us-east-1")
	assert.Contains(defaultAwsRegions, "us-west-2")
	assert.Contains(defaultAwsRegions, "eu-west-1")
	assert.Contains(defaultAwsRegions, "ap-northeast-1")
	assert.Greater(len(defaultAwsRegions), 10)
}

func TestMaxOutputResults(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(50, maxOutputResults)
}

func TestEmptyTarget(t *testing.T) {
	assert := assert.New(t)

	target := &Target{}
	assert.Empty(target.Name)
	assert.Empty(target.PublicDomain)
	assert.Empty(target.PrivateDomain)
}

func TestEmptyUser(t *testing.T) {
	assert := assert.New(t)

	user := &User{}
	assert.Empty(user.Name)
}

func TestEmptyRegion(t *testing.T) {
	assert := assert.New(t)

	region := &Region{}
	assert.Empty(region.Name)
}

func TestAskTeam(t *testing.T)                {}
func TestAskRegion(t *testing.T)              {}
func TestAskTarget(t *testing.T)              {}
func TestCreateStartSession(t *testing.T)     {}
func TestDeleteStartSession(t *testing.T)     {}
func TestSendCommand(t *testing.T)            {}
func TestPrintCommandInvocation(t *testing.T) {}

func TestPrintReady(t *testing.T) {
	// PrintReady just prints to stdout, verify it doesn't panic
	assert.NotPanics(t, func() {
		PrintReady("test-cmd", "us-east-1", "i-1234567890abcdef0")
	})
}
