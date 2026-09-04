package webui

import "testing"

func TestFormatBytes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{-1, "0 B"},
		{1, "1.00 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := formatBytes(tt.in); got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatSpeed(t *testing.T) {
	t.Parallel()
	if got := formatSpeed(1024); got != "1.00 KB/s" {
		t.Errorf("formatSpeed(1024) = %q, want %q", got, "1.00 KB/s")
	}
}

func TestFormatTime(t *testing.T) {
	t.Parallel()
	zero := 0
	neg := -3
	hour := 3661
	done := 90
	tests := []struct {
		name string
		in   *int
		want string
	}{
		{name: "nil", in: nil, want: "Unknown"},
		{name: "negative", in: &neg, want: "Unknown"},
		{name: "zero", in: &zero, want: "Done"},
		{name: "hour", in: &hour, want: "1h 1m"},
		{name: "seconds", in: &done, want: "1m 30s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatTime(tt.in); got != tt.want {
				t.Errorf("formatTime(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestProgressPercent(t *testing.T) {
	t.Parallel()
	if got := progressPercent(50, 100); got != 50 {
		t.Errorf("progressPercent(50, 100) = %v, want 50", got)
	}
	if got := progressPercent(1, 0); got != 0 {
		t.Errorf("progressPercent(1, 0) = %v, want 0", got)
	}
	if got := progressPercent(150, 100); got != 100 {
		t.Errorf("progressPercent(150, 100) = %v, want 100", got)
	}
}
