package session

import (
	"errors"
	"fmt"
	"testing"
)

// Test fixtures for errorChanged (Err* table entries for the lint catalog).
var (
	ErrBoom = errors.New("boom")
	ErrA    = errors.New("a")
	ErrB    = errors.New("b")
)

// stringErr is a distinct error value with a fixed message (avoids ad-hoc errors.New in cases).
type stringErr string

func (e stringErr) Error() string { return string(e) }

func TestErrorChanged(t *testing.T) {
	tests := []struct {
		name string
		a, b error
		want bool
	}{
		{"both nil", nil, nil, false},
		{"nil to error", nil, ErrBoom, true},
		{"error to nil", ErrBoom, nil, true},
		// Distinct instances with the same text compare equal by Error() string.
		{"same message different instances", stringErr("boom"), stringErr("boom"), false},
		{"different messages", ErrA, ErrB, true},
		{"wrapped same text", fmt.Errorf("%w", ErrBoom), ErrBoom, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errorChanged(tt.a, tt.b); got != tt.want {
				t.Fatalf("errorChanged(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
