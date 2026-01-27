package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	// Test command exists and has correct configuration
	assert.NotNil(listCommand)
	assert.Equal("list", listCommand.Use)
	assert.Contains(listCommand.Short, "List")
}

func TestListCommandIsSubcommand(t *testing.T) {
	assert := assert.New(t)

	// Verify list is registered as a subcommand of root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "list" {
			found = true
			break
		}
	}
	assert.True(found, "list command should be registered as subcommand")
}
