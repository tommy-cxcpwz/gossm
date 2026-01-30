package internal

import (
	"testing"
)

func TestValidateInstanceID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "valid 8 hex chars", id: "i-0a1b2c3d", wantErr: false},
		{name: "valid 17 hex chars", id: "i-0a1b2c3d4e5f67890", wantErr: false},
		{name: "valid typical ID", id: "i-0abc123def456789a", wantErr: false},
		{name: "empty string", id: "", wantErr: true},
		{name: "wrong prefix", id: "x-0a1b2c3d", wantErr: true},
		{name: "no prefix", id: "0a1b2c3d", wantErr: true},
		{name: "too short", id: "i-0a1b2c", wantErr: true},
		{name: "too long", id: "i-0a1b2c3d4e5f678901", wantErr: true},
		{name: "invalid chars", id: "i-0a1b2c3g", wantErr: true},
		{name: "uppercase hex", id: "i-0A1B2C3D", wantErr: true},
		{name: "prefix only", id: "i-", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInstanceID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInstanceID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}
