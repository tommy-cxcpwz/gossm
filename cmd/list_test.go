package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tommy-cxcpwz/gossm/internal"
)

func TestFormatFields_AllPopulated_ReturnsValues(t *testing.T) {
	target := &internal.Target{
		TagName:       "web-server",
		PrivateDomain: "ip-10-0-0-1.ec2.internal",
		PublicDomain:  "ec2-1-2-3-4.compute.amazonaws.com",
	}

	name, privateDNS, publicDNS := formatFields(target)

	assert.Equal(t, "web-server", name)
	assert.Equal(t, "ip-10-0-0-1.ec2.internal", privateDNS)
	assert.Equal(t, "ec2-1-2-3-4.compute.amazonaws.com", publicDNS)
}

func TestFormatFields_AllEmpty_ReturnsDashes(t *testing.T) {
	target := &internal.Target{}

	name, privateDNS, publicDNS := formatFields(target)

	assert.Equal(t, "-", name)
	assert.Equal(t, "-", privateDNS)
	assert.Equal(t, "-", publicDNS)
}

func TestFormatFields_PartialEmpty_ReturnsDashForEmpty(t *testing.T) {
	target := &internal.Target{
		TagName:       "my-instance",
		PrivateDomain: "ip-10-0-0-1.ec2.internal",
	}

	name, privateDNS, publicDNS := formatFields(target)

	assert.Equal(t, "my-instance", name)
	assert.Equal(t, "ip-10-0-0-1.ec2.internal", privateDNS)
	assert.Equal(t, "-", publicDNS)
}

func TestListCommand_ShowTagsFlag_Registered(t *testing.T) {
	flag := listCommand.Flags().Lookup("show-tags")

	assert.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
}
