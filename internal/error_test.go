package internal

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapError(t *testing.T) {
	assert := assert.New(t)

	tests := map[string]struct {
		err error
	}{
		"error": {err: fmt.Errorf("[err] obj error")},
	}

	for _, t := range tests {
		err := WrapError(t.err)
		assert.True(errors.Is(err, t.err))
		fmt.Println(err)
	}
}

func TestWrapErrorNil(t *testing.T) {
	assert := assert.New(t)

	// Test with nil error
	result := WrapError(nil)
	assert.Nil(result)
}

func TestErrorVariables(t *testing.T) {
	assert := assert.New(t)

	// Test error variables exist and have expected content
	assert.NotNil(ErrInvalidParams)
	assert.Contains(ErrInvalidParams.Error(), "invalid params")

	assert.NotNil(ErrUnknown)
	assert.Contains(ErrUnknown.Error(), "unknown")
}
