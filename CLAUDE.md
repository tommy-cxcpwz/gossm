# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

gossm is an interactive CLI tool for connecting to AWS EC2 instances via AWS Systems Manager Session Manager. It provides commands for start-session, listing instances, and executing commands remotely.

## Build and Test Commands

```bash
# Build the binary
go build -o gossm .

# Run all tests with race detection and coverage
go test -v $(go list ./... | grep -v vendor) --count 1 -race -coverprofile=coverage.txt -covermode=atomic

# Run tests for a specific package
go test -v ./internal/...
go test -v ./cmd/...

# Lint (requires golangci-lint)
golangci-lint run ./...

# Lint with auto-fix
golangci-lint run --fix ./...
```

## Architecture

### Package Structure

- `main.go` - Entry point, passes version to cmd.Execute()
- `cmd/` - Cobra CLI commands (root, start, exec, list)
- `internal/` - Core business logic for AWS interactions and utilities

### Key Components

**cmd/root.go** - Initializes global credential handling:
- Manages AWS profile selection from flags, env vars, or defaults
- Handles shared credentials files
- Auto-extracts and manages the embedded SSM session-manager-plugin to `~/.gossm/`

**cmd/session.go** - Start session command:
- Starts an interactive SSM session with a selected or specified instance
- Validates instance ID format before connecting

**cmd/exec.go** - Execute command:
- Executes commands on specific instances via SSM Run Command
- Validates instance ID and SSM connectivity before execution

**cmd/list.go** - List command:
- Lists all EC2 instances with SSM agent connected
- Displays in table format with name, ID, and DNS information

**internal/ssm.go** - AWS SSM and EC2 operations:
- `FindInstances()` - Discovers EC2 instances with SSM agent connected (runs SSM and EC2 API calls in parallel for performance)
- `FindInstanceIdsWithConnectedSSM()` - Gets instance IDs with active SSM connections
- `AskTarget()/AskMultiTarget()` - Interactive instance selection via survey library
- `CreateStartSession()/DeleteStartSession()` - SSM session lifecycle
- `SendCommand()` - Sends commands to instances via SSM Run Command
- `PrintCommandInvocation()` - Watches and prints command execution results
- `CallProcess()` - Executes the SSM plugin as a subprocess

**internal/debug.go** - Debug and timing utilities:
- `DebugMode` - Global flag to enable debug output
- `DebugLog()` - Prints debug messages when enabled
- `StartTimer()/Stop()` - Measures and reports operation timing

**internal/assets.go** - Embeds platform-specific session-manager-plugin binaries using `//go:embed assets/*`

**internal/aws.go** - AWS SDK configuration helpers:
- `NewConfig()` - Creates AWS config from explicit credentials or env vars
- `NewSharedConfig()` - Creates AWS config from shared credential files

### Command Flow

1. User runs a command (e.g., `gossm start`)
2. `initConfig()` in root.go loads AWS credentials and region
3. Command handler in respective file calls internal functions
4. `AskTarget()` presents interactive server selection (if needed)
5. `CreateStartSession()` establishes SSM session
6. `CallProcess()` invokes the embedded session-manager-plugin
7. Session terminates and cleanup occurs

### Dependencies

- `spf13/cobra` v1.10.2 - CLI framework
- `spf13/viper` v1.21.0 - Configuration management
- `AlecAivazis/survey/v2` v2.3.7 - Interactive prompts
- `aws/aws-sdk-go-v2` v1.41.1 - AWS SDK for Go v2
- `fatih/color` v1.18.0 - Colored terminal output

Requires Go 1.24+
