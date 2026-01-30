package internal

import (
	"fmt"
	"regexp"
)

var instanceIDRegex = regexp.MustCompile(`^i-[0-9a-f]{8,17}$`)

// ValidateInstanceID validates that the given string is a valid EC2 instance ID.
func ValidateInstanceID(id string) error {
	if !instanceIDRegex.MatchString(id) {
		return fmt.Errorf("invalid instance ID format: %s (must match i-[0-9a-f]{8,17})", id)
	}
	return nil
}
