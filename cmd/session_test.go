package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartSessionCommand(t *testing.T) {
	assert := assert.New(t)

	// Test command exists and has correct configuration
	assert.NotNil(startSessionCommand)
	assert.Equal("start", startSessionCommand.Use)
	assert.Contains(startSessionCommand.Short, "start-session")
}

func TestStartSessionCommandFlags(t *testing.T) {
	assert := assert.New(t)

	// Test target flag exists
	targetFlag := startSessionCommand.Flags().Lookup("target")
	assert.NotNil(targetFlag)
	assert.Equal("t", targetFlag.Shorthand)
	assert.Equal("", targetFlag.DefValue)
}

func TestStartSessionCommandIsSubcommand(t *testing.T) {
	assert := assert.New(t)

	// Verify start is registered as a subcommand of root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "start" {
			found = true
			break
		}
	}
	assert.True(found, "start command should be registered as subcommand")
}
