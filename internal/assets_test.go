package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAsset_ValidKey_ReturnsData(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"DarwinAmd64", "plugin/darwin_amd64/session-manager-plugin"},
		{"DarwinArm64", "plugin/darwin_arm64/session-manager-plugin"},
		{"LinuxAmd64", "plugin/linux_amd64/session-manager-plugin"},
		{"LinuxArm64", "plugin/linux_arm64/session-manager-plugin"},
		{"WindowsAmd64", "plugin/windows_amd64/session-manager-plugin.exe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := GetAsset(tt.key)

			require.NoError(t, err)
			assert.NotEmpty(t, data)
		})
	}
}

func TestGetAsset_InvalidKey_ReturnsError(t *testing.T) {
	_, err := GetAsset("nonexistent-key")

	assert.Error(t, err)
}

func TestGetSsmPluginName_Called_ReturnsPluginFilename(t *testing.T) {
	got := GetSsmPluginName()

	assert.NotEmpty(t, got)
}

func TestGetSsmPlugin_Called_ReturnsNonEmptyData(t *testing.T) {
	data, err := GetSsmPlugin()

	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestGetSsmPluginSize_Called_ReturnsPositiveSize(t *testing.T) {
	size, err := GetSsmPluginSize()

	require.NoError(t, err)
	assert.Greater(t, size, int64(0))
}

func TestGetSSMPluginKey_Called_ReturnsValidPath(t *testing.T) {
	key := getSSMPluginKey()

	assert.NotEmpty(t, key)
}
