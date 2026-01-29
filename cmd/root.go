package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tommy-cxcpwz/gossm/internal"
)

const (
	_defaultProfile   = "default"
	_credentialFormat = "[%s]\naws_access_key_id = %s\naws_secret_access_key = %s\naws_session_token = %s\n"
)

var (
	// rootCmd represents the base command when called without any sub-commands
	rootCmd = &cobra.Command{
		Use:   "gossm",
		Short: `gossm is interactive CLI tool that you select server in AWS and then could connect using AWS Systems Manager Session Manager.`,
		Long:  `gossm is interactive CLI tool that you select server in AWS and then could connect using AWS Systems Manager Session Manager.`,
	}

	_credential              *Credential
	_credentialWithTemporary = fmt.Sprintf("%s_temporary", config.DefaultSharedCredentialsFilename())
)

type Credential struct {
	awsProfile    string
	awsConfig     *aws.Config
	gossmHomePath string
	ssmPluginPath string
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		panicRed(err)
	}
}

// panicRed raises error with text.
func panicRed(err error) {
	fmt.Println(color.RedString("[err] %s", err.Error()))
	os.Exit(1)
}

// resolveAWSProfile determines the AWS profile to use based on flag, environment variable, or default.
func resolveAWSProfile(flagProfile string) string {
	if flagProfile != "" {
		return flagProfile
	}
	if envProfile := os.Getenv("AWS_PROFILE"); envProfile != "" {
		return envProfile
	}
	return _defaultProfile
}

// checkPluginNeedsUpdate checks if the SSM plugin at the given path needs to be updated.
// Returns true if the plugin doesn't exist or has a different size than the embedded plugin.
func checkPluginNeedsUpdate(pluginPath string, getEmbeddedSize func() (int64, error)) (bool, error) {
	info, err := os.Stat(pluginPath)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	embeddedSize, err := getEmbeddedSize()
	if err != nil {
		return false, err
	}

	return info.Size() != embeddedSize, nil
}

// getGossmHomePath returns the path to the gossm home directory.
func getGossmHomePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gossm"), nil
}

// ensureDirectoryExists creates the directory if it doesn't exist.
func ensureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0700) // Secure permissions: owner only
	}
	return nil
}

// resolveSharedCredentialFile validates the AWS_SHARED_CREDENTIALS_FILE env var.
// Returns the cleaned absolute path if valid, or empty string if invalid/unset.
func resolveSharedCredentialFile() string {
	sharedCredFile := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if sharedCredFile == "" {
		return ""
	}
	absPath, err := filepath.Abs(sharedCredFile)
	if err != nil {
		color.Yellow("[Warning] invalid AWS_SHARED_CREDENTIALS_FILE environments path: %v", err)
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		return ""
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		color.Yellow("[Warning] not found AWS_SHARED_CREDENTIALS_FILE environments file, such as %s", absPath)
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		return ""
	}
	return absPath
}

// formatTemporaryCredentials formats credentials into the INI file format.
func formatTemporaryCredentials(profile, accessKey, secretKey, sessionToken string) string {
	return fmt.Sprintf(_credentialFormat, profile, accessKey, secretKey, sessionToken)
}

// isCredentialValid checks if AWS credentials are valid (not expired and have required fields).
func isCredentialValid(cred aws.Credentials) bool {
	return !cred.Expired() && cred.AccessKeyID != "" && cred.SecretAccessKey != ""
}

// writeTemporaryCredentialFile writes credential data to the temporary file and sets the env var.
func writeTemporaryCredentialFile(profile string, cred aws.Credentials) error {
	data := formatTemporaryCredentials(profile, cred.AccessKeyID, cred.SecretAccessKey, cred.SessionToken)
	if err := os.WriteFile(_credentialWithTemporary, []byte(data), 0600); err != nil {
		return err
	}
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", _credentialWithTemporary)
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Enable debug mode if flag is set
	internal.DebugMode = viper.GetBool("debug")
	timer := internal.StartTimer("initConfig")
	defer timer.Stop()

	_credential = &Credential{}
	// 1. get aws profile
	_credential.awsProfile = resolveAWSProfile(viper.GetString("profile"))

	// 2. get region
	awsRegion := viper.GetString("region")

	// 3. update or create aws ssm plugin.
	gossmHomePath, err := getGossmHomePath()
	if err != nil {
		panicRed(internal.WrapError(err))
	}
	_credential.gossmHomePath = gossmHomePath

	if err := ensureDirectoryExists(_credential.gossmHomePath); err != nil {
		panicRed(internal.WrapError(err))
	}

	_credential.ssmPluginPath = filepath.Join(_credential.gossmHomePath, internal.GetSsmPluginName())

	// Check if plugin needs to be created/updated (compare sizes first to avoid loading large binary)
	pluginTimer := internal.StartTimer("check SSM plugin")
	needsUpdate, err := checkPluginNeedsUpdate(_credential.ssmPluginPath, internal.GetSsmPluginSize)
	if err != nil {
		panicRed(internal.WrapError(err))
	}
	var info os.FileInfo
	if !needsUpdate {
		info, _ = os.Stat(_credential.ssmPluginPath)
	}

	if needsUpdate {
		internal.DebugLog("plugin needs update, loading binary...")
		plugin, err := internal.GetSsmPlugin()
		if err != nil {
			panicRed(internal.WrapError(err))
		}
		if info == nil {
			color.Green("[create] aws ssm plugin")
		} else {
			color.Green("[update] aws ssm plugin")
		}
		if err := os.WriteFile(_credential.ssmPluginPath, plugin, 0755); err != nil {
			panicRed(internal.WrapError(err))
		}
	}
	pluginTimer.Stop()

	// 4. set shared credential.
	sharedCredFile := resolveSharedCredentialFile()

	// if shared cred file is exist.
	if sharedCredFile != "" {
		credTimer := internal.StartTimer("load shared credentials")
		awsConfig, err := internal.NewSharedConfig(context.Background(),
			_credential.awsProfile,
			[]string{config.DefaultSharedConfigFilename()},
			[]string{sharedCredFile},
		)
		if err != nil {
			panicRed(internal.WrapError(err))
		}

		cred, err := awsConfig.Credentials.Retrieve(context.Background())
		// delete invalid shared credential.
		if err != nil || !isCredentialValid(cred) {
			color.Yellow("[Expire] credential file %s", sharedCredFile)
			os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
		} else {
			_credential.awsConfig = &awsConfig
		}
		credTimer.Stop()
	}

	if _credential.awsConfig == nil { // not use shared credential
		credLoadTimer := internal.StartTimer("load credentials")
		var temporaryCredentials aws.Credentials
		var temporaryConfig aws.Config

		// Load credentials directly from config + credentials file (skip config-only attempt)
		internal.DebugLog("loading from config + credentials file")
		temporaryConfig, err = internal.NewSharedConfig(context.Background(), _credential.awsProfile,
			[]string{config.DefaultSharedConfigFilename()}, []string{config.DefaultSharedCredentialsFilename()})
		if err != nil {
			panicRed(internal.WrapError(err))
		}

		retrieveTimer := internal.StartTimer("retrieve credentials")
		temporaryCredentials, err = temporaryConfig.Credentials.Retrieve(context.Background())
		retrieveTimer.Stop()
		if err != nil {
			panicRed(internal.WrapError(err))
		}

		if !isCredentialValid(temporaryCredentials) {
			panicRed(internal.WrapError(fmt.Errorf("[err] not found valid credentials")))
		}

		// extract aws region if awsRegion is empty.
		if awsRegion == "" {
			awsRegion = temporaryConfig.Region
		}

		// [ISSUE] KMS Encrypt, must use AWS_SHARED_CREDENTIALS_FILE with SharedConfig.
		// [INFO] write temporaryCredentials to file.
		if err := writeTemporaryCredentialFile(_credential.awsProfile, temporaryCredentials); err != nil {
			panicRed(internal.WrapError(err))
		}

		finalConfigTimer := internal.StartTimer("create final AWS config")
		awsConfig, err := internal.NewSharedConfig(context.Background(),
			_credential.awsProfile, []string{}, []string{_credentialWithTemporary},
		)
		if err != nil {
			panicRed(internal.WrapError(err))
		}
		_credential.awsConfig = &awsConfig
		finalConfigTimer.Stop()
		credLoadTimer.Stop()
	}

	// set region
	if awsRegion != "" {
		_credential.awsConfig.Region = awsRegion
	}
	if _credential.awsConfig.Region == "" { // ask region
		askRegion, err := internal.AskRegion(context.Background(), *_credential.awsConfig)
		if err != nil {
			panicRed(internal.WrapError(err))
		}
		_credential.awsConfig.Region = askRegion.Name
	}
	color.Green("region (%s)", _credential.awsConfig.Region)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringP("profile", "p", "", `[optional] if you are having multiple aws profiles, it is one of profiles (default is AWS_PROFILE environment variable or default)`)
	rootCmd.PersistentFlags().StringP("region", "r", "", `[optional] it is region in AWS that would like to do something`)
	rootCmd.PersistentFlags().Bool("debug", false, `[optional] enable debug mode to show timing information`)

	// set version flag
	rootCmd.InitDefaultVersionFlag()

	// mapping viper
	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}
