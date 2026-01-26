package internal

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

var DebugMode bool

// DebugLog prints a debug message if debug mode is enabled.
func DebugLog(format string, args ...interface{}) {
	if DebugMode {
		msg := fmt.Sprintf(format, args...)
		fmt.Println(color.MagentaString("[debug] %s", msg))
	}
}

// DebugTimer is a helper for timing operations.
type DebugTimer struct {
	name  string
	start time.Time
}

// StartTimer starts a new timer with the given operation name.
func StartTimer(name string) *DebugTimer {
	if DebugMode {
		fmt.Println(color.MagentaString("[debug] starting: %s", name))
	}
	return &DebugTimer{
		name:  name,
		start: time.Now(),
	}
}

// Stop stops the timer and prints the elapsed time if debug mode is enabled.
func (t *DebugTimer) Stop() time.Duration {
	elapsed := time.Since(t.start)
	if DebugMode {
		fmt.Println(color.MagentaString("[debug] completed: %s (took %v)", t.name, elapsed.Round(time.Millisecond)))
	}
	return elapsed
}
