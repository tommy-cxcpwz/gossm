package internal

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapError_NonNilError_PreservesOriginalError(t *testing.T) {
	original := fmt.Errorf("something failed")

	wrapped := WrapError(original)

	require.NotNil(t, wrapped)
	assert.True(t, errors.Is(wrapped, original))
}

func TestWrapError_NilError_ReturnsNil(t *testing.T) {
	result := WrapError(nil)

	assert.Nil(t, result)
}
