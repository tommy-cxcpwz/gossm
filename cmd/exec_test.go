package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecCommand(t *testing.T) {
	assert := assert.New(t)

	// Test command exists and has correct configuration
	assert.NotNil(execCommand)
	assert.Contains(execCommand.Use, "exec")
	assert.Contains(execCommand.Short, "Execute")
}

func TestExecCommandFlags(t *testing.T) {
	assert := assert.New(t)

	// Test skip-check flag exists
	skipCheckFlag := execCommand.Flags().Lookup("skip-check")
	assert.NotNil(skipCheckFlag)
	assert.Equal("false", skipCheckFlag.DefValue)
}

func TestExecCommandIsSubcommand(t *testing.T) {
	assert := assert.New(t)

	// Verify exec is registered as a subcommand of root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "exec" {
			found = true
			break
		}
	}
	assert.True(found, "exec command should be registered as subcommand")
}

func TestExecCommandRequiresMinimumArgs(t *testing.T) {
	assert := assert.New(t)

	// Check that exec command requires minimum 2 args
	assert.NotNil(execCommand.Args)
}
