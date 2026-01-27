package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebugLog(t *testing.T) {
	// Test with debug mode disabled
	DebugMode = false
	assert.NotPanics(t, func() {
		DebugLog("test message %s", "arg")
	})

	// Test with debug mode enabled
	DebugMode = true
	assert.NotPanics(t, func() {
		DebugLog("test message %s", "arg")
	})

	// Reset debug mode
	DebugMode = false
}

func TestStartTimer(t *testing.T) {
	assert := assert.New(t)

	// Test with debug mode disabled
	DebugMode = false
	timer := StartTimer("test-operation")
	assert.NotNil(timer)
	assert.Equal("test-operation", timer.name)
	assert.False(timer.start.IsZero())

	// Test with debug mode enabled
	DebugMode = true
	timer2 := StartTimer("test-operation-2")
	assert.NotNil(timer2)
	assert.Equal("test-operation-2", timer2.name)

	// Reset debug mode
	DebugMode = false
}

func TestDebugTimerStop(t *testing.T) {
	assert := assert.New(t)

	// Test with debug mode disabled
	DebugMode = false
	timer := StartTimer("test-stop")
	time.Sleep(10 * time.Millisecond)
	elapsed := timer.Stop()
	assert.GreaterOrEqual(elapsed, 10*time.Millisecond)

	// Test with debug mode enabled
	DebugMode = true
	timer2 := StartTimer("test-stop-2")
	time.Sleep(10 * time.Millisecond)
	elapsed2 := timer2.Stop()
	assert.GreaterOrEqual(elapsed2, 10*time.Millisecond)

	// Reset debug mode
	DebugMode = false
}

func TestDebugTimerStruct(t *testing.T) {
	assert := assert.New(t)

	timer := &DebugTimer{
		name:  "test",
		start: time.Now(),
	}

	assert.Equal("test", timer.name)
	assert.False(timer.start.IsZero())
}
