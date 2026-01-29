package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDebugLog_DebugDisabled_DoesNotPanic(t *testing.T) {
	DebugMode = false
	defer func() { DebugMode = false }()

	assert.NotPanics(t, func() {
		DebugLog("test message %s", "arg")
	})
}

func TestDebugLog_DebugEnabled_DoesNotPanic(t *testing.T) {
	DebugMode = true
	defer func() { DebugMode = false }()

	assert.NotPanics(t, func() {
		DebugLog("test message %s", "arg")
	})
}

func TestStartTimer_Called_ReturnsNonNilTimer(t *testing.T) {
	DebugMode = false
	defer func() { DebugMode = false }()

	timer := StartTimer("test-operation")

	require.NotNil(t, timer)
}

func TestStartTimer_DebugEnabled_ReturnsNonNilTimer(t *testing.T) {
	DebugMode = true
	defer func() { DebugMode = false }()

	timer := StartTimer("test-operation")

	require.NotNil(t, timer)
}

func TestDebugTimerStop_Called_ReturnsPositiveDuration(t *testing.T) {
	DebugMode = false
	defer func() { DebugMode = false }()

	timer := StartTimer("test-stop")

	elapsed := timer.Stop()

	assert.GreaterOrEqual(t, elapsed.Nanoseconds(), int64(0))
}

func TestDebugTimerStop_DebugEnabled_ReturnsPositiveDuration(t *testing.T) {
	DebugMode = true
	defer func() { DebugMode = false }()

	timer := StartTimer("test-stop")

	elapsed := timer.Stop()

	assert.GreaterOrEqual(t, elapsed.Nanoseconds(), int64(0))
}
