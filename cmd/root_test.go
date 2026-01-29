package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- test helpers ---

// saveEnv saves an environment variable and returns a restore function.
func saveEnv(t *testing.T, key string) func() {
	t.Helper()
	orig, ok := os.LookupEnv(key)
	return func() {
		if ok {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	}
}

// createTempFile creates a temporary file with the given content and returns its path.
// The file is automatically cleaned up when the test finishes.
func createTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "gossm-test-*")
	require.NoError(t, err)
	if content != "" {
		_, err = f.WriteString(content)
		require.NoError(t, err)
	}
	require.NoError(t, f.Close())
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

// --- resolveAWSProfile ---

func TestResolveAWSProfile_FlagProvided_ReturnsFlagValue(t *testing.T) {
	restore := saveEnv(t, "AWS_PROFILE")
	defer restore()
	os.Setenv("AWS_PROFILE", "env-profile")

	got := resolveAWSProfile("flag-profile")

	assert.Equal(t, "flag-profile", got)
}

func TestResolveAWSProfile_NoFlagWithEnvSet_ReturnsEnvVar(t *testing.T) {
	restore := saveEnv(t, "AWS_PROFILE")
	defer restore()
	os.Setenv("AWS_PROFILE", "env-profile")

	got := resolveAWSProfile("")

	assert.Equal(t, "env-profile", got)
}

func TestResolveAWSProfile_NothingSet_ReturnsDefault(t *testing.T) {
	restore := saveEnv(t, "AWS_PROFILE")
	defer restore()
	os.Unsetenv("AWS_PROFILE")

	got := resolveAWSProfile("")

	assert.Equal(t, "default", got)
}

// --- checkPluginNeedsUpdate ---

func TestCheckPluginNeedsUpdate_FileNotExists_ReturnsTrue(t *testing.T) {
	stubSize := func() (int64, error) { return 100, nil }

	needsUpdate, err := checkPluginNeedsUpdate("/nonexistent/path/plugin", stubSize)

	require.NoError(t, err)
	assert.True(t, needsUpdate)
}

func TestCheckPluginNeedsUpdate_SameSize_ReturnsFalse(t *testing.T) {
	path := createTempFile(t, "hello")
	stubSize := func() (int64, error) { return 5, nil } // "hello" = 5 bytes

	needsUpdate, err := checkPluginNeedsUpdate(path, stubSize)

	require.NoError(t, err)
	assert.False(t, needsUpdate)
}

func TestCheckPluginNeedsUpdate_DifferentSize_ReturnsTrue(t *testing.T) {
	path := createTempFile(t, "hello")
	stubSize := func() (int64, error) { return 999, nil }

	needsUpdate, err := checkPluginNeedsUpdate(path, stubSize)

	require.NoError(t, err)
	assert.True(t, needsUpdate)
}

func TestCheckPluginNeedsUpdate_SizeCallFails_ReturnsError(t *testing.T) {
	path := createTempFile(t, "hello")
	stubSize := func() (int64, error) { return 0, fmt.Errorf("embed error") }

	_, err := checkPluginNeedsUpdate(path, stubSize)

	assert.Error(t, err)
}

func TestCheckPluginNeedsUpdate_SymlinkLoop_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	linkPath := filepath.Join(tmpDir, "loop")
	os.Symlink(linkPath, linkPath)
	stubSize := func() (int64, error) { return 100, nil }

	needsUpdate, err := checkPluginNeedsUpdate(linkPath, stubSize)

	if err != nil {
		assert.False(t, needsUpdate)
	}
}

// --- getGossmHomePath ---

func TestGetGossmHomePath_Success_ReturnsPathEndingWithDotGossm(t *testing.T) {
	path, err := getGossmHomePath()

	require.NoError(t, err)
	assert.Equal(t, ".gossm", filepath.Base(path))
}

// --- ensureDirectoryExists ---

func TestEnsureDirectoryExists_AlreadyExists_ReturnsNoError(t *testing.T) {
	dir := t.TempDir()

	err := ensureDirectoryExists(dir)

	assert.NoError(t, err)
}

func TestEnsureDirectoryExists_NotExists_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "a", "b", "c")

	err := ensureDirectoryExists(dir)

	require.NoError(t, err)
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// --- resolveSharedCredentialFile ---

func TestResolveSharedCredentialFile_EnvUnset_ReturnsEmpty(t *testing.T) {
	restore := saveEnv(t, "AWS_SHARED_CREDENTIALS_FILE")
	defer restore()
	os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")

	got := resolveSharedCredentialFile()

	assert.Equal(t, "", got)
}

func TestResolveSharedCredentialFile_FileNotExists_UnsetsEnvAndReturnsEmpty(t *testing.T) {
	restore := saveEnv(t, "AWS_SHARED_CREDENTIALS_FILE")
	defer restore()
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "/tmp/nonexistent_cred_file_xyz")

	got := resolveSharedCredentialFile()

	assert.Equal(t, "", got)
	_, envSet := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE")
	assert.False(t, envSet)
}

func TestResolveSharedCredentialFile_ValidFile_ReturnsAbsolutePath(t *testing.T) {
	restore := saveEnv(t, "AWS_SHARED_CREDENTIALS_FILE")
	defer restore()
	path := createTempFile(t, "")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", path)

	got := resolveSharedCredentialFile()

	absPath, _ := filepath.Abs(path)
	assert.Equal(t, absPath, got)
}

// --- isCredentialValid ---

func TestIsCredentialValid(t *testing.T) {
	tests := []struct {
		name string
		cred aws.Credentials
		want bool
	}{
		{
			name: "BothFieldsPresent_ReturnsTrue",
			cred: aws.Credentials{AccessKeyID: "AKID", SecretAccessKey: "SECRET"},
			want: true,
		},
		{
			name: "MissingAccessKey_ReturnsFalse",
			cred: aws.Credentials{AccessKeyID: "", SecretAccessKey: "SECRET"},
			want: false,
		},
		{
			name: "MissingSecretKey_ReturnsFalse",
			cred: aws.Credentials{AccessKeyID: "AKID", SecretAccessKey: ""},
			want: false,
		},
		{
			name: "ZeroValue_ReturnsFalse",
			cred: aws.Credentials{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCredentialValid(tt.cred)

			assert.Equal(t, tt.want, got)
		})
	}
}

// --- writeTemporaryCredentialFile ---

func TestWriteTemporaryCredentialFile_ValidPath_WritesFileAndSetsEnv(t *testing.T) {
	origTempPath := _credentialWithTemporary
	restoreEnv := saveEnv(t, "AWS_SHARED_CREDENTIALS_FILE")
	defer func() {
		_credentialWithTemporary = origTempPath
		restoreEnv()
	}()

	tmpPath := createTempFile(t, "")
	_credentialWithTemporary = tmpPath
	cred := aws.Credentials{
		AccessKeyID:     "AKID123",
		SecretAccessKey: "SECRET456",
		SessionToken:    "TOKEN789",
	}

	err := writeTemporaryCredentialFile("myprofile", cred)

	require.NoError(t, err)
	data, err := os.ReadFile(tmpPath)
	require.NoError(t, err)
	want := formatTemporaryCredentials("myprofile", "AKID123", "SECRET456", "TOKEN789")
	assert.Equal(t, want, string(data))
	assert.Equal(t, tmpPath, os.Getenv("AWS_SHARED_CREDENTIALS_FILE"))
}

func TestWriteTemporaryCredentialFile_InvalidPath_ReturnsError(t *testing.T) {
	origTempPath := _credentialWithTemporary
	defer func() { _credentialWithTemporary = origTempPath }()
	_credentialWithTemporary = "/nonexistent/dir/file"

	err := writeTemporaryCredentialFile("p", aws.Credentials{})

	assert.Error(t, err)
}

// --- Execute ---

func TestExecute_VersionFlag_ExitsWithoutError(t *testing.T) {
	originalArgs := os.Args
	originalVersion := rootCmd.Version
	defer func() {
		os.Args = originalArgs
		rootCmd.Version = originalVersion
		rootCmd.SetArgs([]string{})
	}()

	os.Args = []string{"gossm", "--version"}
	rootCmd.SetArgs([]string{"--version"})

	Execute("1.0.0-test")
}

// --- panicRed ---

func TestPanicRed_Called_ExitsWithNonZeroCode(t *testing.T) {
	if os.Getenv("TEST_PANIC_RED") == "1" {
		panicRed(fmt.Errorf("test error"))
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestPanicRed_Called_ExitsWithNonZeroCode")
	cmd.Env = append(os.Environ(), "TEST_PANIC_RED=1")

	err := cmd.Run()

	var exitErr *exec.ExitError
	assert.ErrorAs(t, err, &exitErr)
	assert.False(t, exitErr.Success())
}
