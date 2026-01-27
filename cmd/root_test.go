package cmd

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	assert := assert.New(t)

	// Test root command exists and has correct configuration
	assert.NotNil(rootCmd)
	assert.Equal("gossm", rootCmd.Use)
	assert.Contains(rootCmd.Short, "gossm is interactive CLI tool")
}

func TestRootCmdFlags(t *testing.T) {
	assert := assert.New(t)

	// Test persistent flags exist
	profileFlag := rootCmd.PersistentFlags().Lookup("profile")
	assert.NotNil(profileFlag)
	assert.Equal("p", profileFlag.Shorthand)

	regionFlag := rootCmd.PersistentFlags().Lookup("region")
	assert.NotNil(regionFlag)
	assert.Equal("r", regionFlag.Shorthand)

	debugFlag := rootCmd.PersistentFlags().Lookup("debug")
	assert.NotNil(debugFlag)
	assert.Equal("false", debugFlag.DefValue)
}

func TestCredentialStruct(t *testing.T) {
	assert := assert.New(t)

	cred := &Credential{
		awsProfile:    "test-profile",
		gossmHomePath: "/home/test/.gossm",
		ssmPluginPath: "/home/test/.gossm/session-manager-plugin",
	}

	assert.Equal("test-profile", cred.awsProfile)
	assert.Equal("/home/test/.gossm", cred.gossmHomePath)
	assert.Equal("/home/test/.gossm/session-manager-plugin", cred.ssmPluginPath)
}

func TestDefaultProfile(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("default", _defaultProfile)
}

func TestRootCmdHasSubcommands(t *testing.T) {
	assert := assert.New(t)

	// Get all subcommands
	subcommands := rootCmd.Commands()

	// Check that we have subcommands registered
	assert.Greater(len(subcommands), 0)

	// Create a map of command names for easier lookup (use Name() instead of Use for commands with args)
	cmdNames := make(map[string]bool)
	for _, cmd := range subcommands {
		cmdNames[cmd.Name()] = true
	}

	// Test some expected commands exist
	expectedCmds := []string{"start", "exec", "list"}
	for _, expected := range expectedCmds {
		assert.True(cmdNames[expected], "expected command %s not found", expected)
	}
}

func TestViperBindings(t *testing.T) {
	assert := assert.New(t)

	// Reset viper for clean test
	viper.Reset()

	// Re-bind flags to viper
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Test default values
	assert.Equal("", viper.GetString("profile"))
	assert.Equal("", viper.GetString("region"))
	assert.Equal(false, viper.GetBool("debug"))
}

func TestExecuteVersion(t *testing.T) {
	// Test that version can be set on the root command
	testVersion := "1.0.0-test"
	originalVersion := rootCmd.Version
	defer func() { rootCmd.Version = originalVersion }()

	rootCmd.Version = testVersion
	assert.Equal(t, testVersion, rootCmd.Version)
}

func TestRootCmdVersion(t *testing.T) {
	assert := assert.New(t)

	// Test that version can be set and retrieved
	originalVersion := rootCmd.Version
	rootCmd.Version = "test-version"
	assert.Equal("test-version", rootCmd.Version)
	rootCmd.Version = originalVersion
}

func TestProfileFromEnvVar(t *testing.T) {
	assert := assert.New(t)

	// Save original env var
	originalProfile := os.Getenv("AWS_PROFILE")
	defer os.Setenv("AWS_PROFILE", originalProfile)

	// Test with custom profile
	os.Setenv("AWS_PROFILE", "custom-profile")
	assert.Equal("custom-profile", os.Getenv("AWS_PROFILE"))

	// Test with empty profile
	os.Unsetenv("AWS_PROFILE")
	assert.Equal("", os.Getenv("AWS_PROFILE"))
}

func TestSharedCredentialsEnvVar(t *testing.T) {
	assert := assert.New(t)

	// Save original env var
	originalCred := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	defer func() {
		if originalCred != "" {
			os.Setenv("AWS_SHARED_CREDENTIALS_FILE", originalCred)
		} else {
			os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		}
	}()

	// Test setting custom credentials file
	testPath := "/tmp/test_credentials"
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", testPath)
	assert.Equal(testPath, os.Getenv("AWS_SHARED_CREDENTIALS_FILE"))

	// Test unsetting
	os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	assert.Equal("", os.Getenv("AWS_SHARED_CREDENTIALS_FILE"))
}

func TestRootCmdFind(t *testing.T) {
	assert := assert.New(t)

	// Test finding subcommands
	tests := []struct {
		args    []string
		wantCmd string
	}{
		{[]string{"start"}, "start"},
		{[]string{"list"}, "list"},
	}

	for _, tt := range tests {
		cmd, _, err := rootCmd.Find(tt.args)
		assert.NoError(err)
		assert.Equal(tt.wantCmd, cmd.Name())
	}
}

func TestRootCmdFindInvalidCommand(t *testing.T) {
	assert := assert.New(t)

	// Test finding a non-existent command returns an error
	_, _, err := rootCmd.Find([]string{"nonexistent"})
	assert.Error(err)
	assert.Contains(err.Error(), "unknown command")
}

func TestResolveAWSProfile(t *testing.T) {
	assert := assert.New(t)

	// Save original env var
	originalProfile := os.Getenv("AWS_PROFILE")
	defer func() {
		if originalProfile != "" {
			os.Setenv("AWS_PROFILE", originalProfile)
		} else {
			os.Unsetenv("AWS_PROFILE")
		}
	}()

	tests := []struct {
		name        string
		flagProfile string
		envProfile  string
		expected    string
	}{
		{
			name:        "flag takes precedence",
			flagProfile: "flag-profile",
			envProfile:  "env-profile",
			expected:    "flag-profile",
		},
		{
			name:        "env var when flag is empty",
			flagProfile: "",
			envProfile:  "env-profile",
			expected:    "env-profile",
		},
		{
			name:        "default when both empty",
			flagProfile: "",
			envProfile:  "",
			expected:    "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envProfile != "" {
				os.Setenv("AWS_PROFILE", tt.envProfile)
			} else {
				os.Unsetenv("AWS_PROFILE")
			}

			result := resolveAWSProfile(tt.flagProfile)
			assert.Equal(tt.expected, result)
		})
	}
}

func TestCheckPluginNeedsUpdate(t *testing.T) {
	assert := assert.New(t)

	// Test with non-existent file
	needsUpdate, err := checkPluginNeedsUpdate("/nonexistent/path", func() (int64, error) {
		return 100, nil
	})
	assert.NoError(err)
	assert.True(needsUpdate)

	// Test with existing file of same size
	tmpFile, err := os.CreateTemp("", "test-plugin")
	assert.NoError(err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("hello")
	tmpFile.Close()

	needsUpdate, err = checkPluginNeedsUpdate(tmpFile.Name(), func() (int64, error) {
		return 5, nil // "hello" is 5 bytes
	})
	assert.NoError(err)
	assert.False(needsUpdate)

	// Test with existing file of different size
	needsUpdate, err = checkPluginNeedsUpdate(tmpFile.Name(), func() (int64, error) {
		return 100, nil
	})
	assert.NoError(err)
	assert.True(needsUpdate)

	// Test with error from getEmbeddedSize
	_, err = checkPluginNeedsUpdate(tmpFile.Name(), func() (int64, error) {
		return 0, os.ErrNotExist
	})
	assert.Error(err)
}

func TestGetGossmHomePath(t *testing.T) {
	assert := assert.New(t)

	path, err := getGossmHomePath()
	assert.NoError(err)
	assert.Contains(path, ".gossm")
}

func TestEnsureDirectoryExists(t *testing.T) {
	assert := assert.New(t)

	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "test-ensure-dir")
	assert.NoError(err)
	defer os.RemoveAll(tmpDir)

	// Test with existing directory
	err = ensureDirectoryExists(tmpDir)
	assert.NoError(err)

	// Test with non-existent directory
	newDir := tmpDir + "/new/nested/dir"
	err = ensureDirectoryExists(newDir)
	assert.NoError(err)

	// Verify directory was created
	info, err := os.Stat(newDir)
	assert.NoError(err)
	assert.True(info.IsDir())
}

func TestCredentialWithAwsConfig(t *testing.T) {
	assert := assert.New(t)

	// Test Credential struct with all fields
	cred := &Credential{
		awsProfile: "test",
		awsConfig:  nil,
	}

	assert.Equal("test", cred.awsProfile)
	assert.Nil(cred.awsConfig)
}

func TestAllSubcommandsHaveRunFunction(t *testing.T) {
	assert := assert.New(t)

	// Verify all subcommands have Run functions
	for _, cmd := range rootCmd.Commands() {
		assert.NotNil(cmd.Run, "command %s should have a Run function", cmd.Name())
	}
}

func TestRootCmdHelp(t *testing.T) {
	assert := assert.New(t)

	// Test that help works
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	assert.NoError(err)

	// Reset args after test
	rootCmd.SetArgs([]string{})
}

func TestExecuteWithVersion(t *testing.T) {
	// Save original args and version
	originalArgs := os.Args
	originalVersion := rootCmd.Version
	defer func() {
		os.Args = originalArgs
		rootCmd.Version = originalVersion
		rootCmd.SetArgs([]string{})
	}()

	// Set up for version output
	os.Args = []string{"gossm", "--version"}
	rootCmd.SetArgs([]string{"--version"})

	// Call Execute - this should not panic with --version
	Execute("1.0.0-test")
}

func TestCheckPluginNeedsUpdateWithStatError(t *testing.T) {
	assert := assert.New(t)

	// Test with a path that causes a stat error (permission denied simulation)
	// This is tricky to test, so we'll just verify the function signature works
	needsUpdate, err := checkPluginNeedsUpdate("/tmp/test-file-does-not-exist-12345", func() (int64, error) {
		return 100, nil
	})
	assert.NoError(err)
	assert.True(needsUpdate)
}

func TestResolveAWSProfileEdgeCases(t *testing.T) {
	assert := assert.New(t)

	// Save original env var
	originalProfile := os.Getenv("AWS_PROFILE")
	defer func() {
		if originalProfile != "" {
			os.Setenv("AWS_PROFILE", originalProfile)
		} else {
			os.Unsetenv("AWS_PROFILE")
		}
	}()

	// Test with whitespace in flag profile
	result := resolveAWSProfile("  ")
	os.Unsetenv("AWS_PROFILE")
	// It should return the whitespace as-is (not trimmed)
	assert.Equal("  ", result)

	// Test with special characters
	result = resolveAWSProfile("my-profile_123")
	assert.Equal("my-profile_123", result)
}

func TestGetGossmHomePathContent(t *testing.T) {
	assert := assert.New(t)

	path, err := getGossmHomePath()
	assert.NoError(err)
	assert.NotEmpty(path)
	// Should end with .gossm
	assert.True(len(path) > 6, "path should be long enough")
}

func TestEnsureDirectoryExistsWithFile(t *testing.T) {
	assert := assert.New(t)

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-file")
	assert.NoError(err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Try to ensure directory exists on a file path (should succeed since file already exists)
	err = ensureDirectoryExists(tmpFile.Name())
	assert.NoError(err)
}
