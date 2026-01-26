package internal

import (
	"embed"
	"fmt"
	"runtime"
	"strings"
)

//go:embed assets/*
var assets embed.FS

// GetAsset returns asset file.
// cannot be accessed from outer package.
func GetAsset(filename string) ([]byte, error) {
	return assets.ReadFile("assets/" + filename)
}

// GetSsmPluginName returns filename for aws ssm plugin.
func GetSsmPluginName() string {
	if strings.ToLower(runtime.GOOS) == "windows" {
		return "session-manager-plugin.exe"
	} else {
		return "session-manager-plugin"
	}
}

// GetSsmPlugin returns the aws ssm plugin binary.
func GetSsmPlugin() ([]byte, error) {
	return GetAsset(getSSMPluginKey())
}

// GetSsmPluginSize returns the size of the embedded ssm plugin without reading it.
func GetSsmPluginSize() (int64, error) {
	f, err := assets.Open("assets/" + getSSMPluginKey())
	if err != nil {
		return 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func getSSMPluginKey() string {
	return fmt.Sprintf("plugin/%s_%s/%s",
		strings.ToLower(runtime.GOOS), strings.ToLower(runtime.GOARCH), GetSsmPluginName())
}
