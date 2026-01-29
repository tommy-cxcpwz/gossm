package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
)

var mockProfile string

func TestMain(m *testing.M) {
	if os.Getenv("CIRCLECI") != "" {
		os.Exit(0)
	}

	mockProfile = "default"
	filename := filepath.Join(os.Getenv("HOME"), ".aws/credentials")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.Exit(0)
	}

	_, err := NewSharedConfig(context.Background(), mockProfile,
		[]string{config.DefaultSharedConfigFilename()},
		[]string{config.DefaultSharedCredentialsFilename()},
	)
	if err != nil {
		os.Exit(0)
	}

	os.Exit(m.Run())
}
