package webui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

var sizeUnits = [...]string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

// formatBytes renders a byte count as a short binary unit string.
func formatBytes(n int64) string {
	if n <= 0 {
		return "0 B"
	}
	fn := float64(n)
	exp := int(math.Log(fn) / math.Log(1024))
	if exp >= len(sizeUnits) {
		exp = len(sizeUnits) - 1
	}
	if exp < 0 {
		exp = 0
	}
	val := fn / math.Pow(1024, float64(exp))
	return fmt.Sprintf("%.2f %s", val, sizeUnits[exp])
}

// formatSpeed renders a bytes-per-second rate.
func formatSpeed(bytesPerSec int) string {
	return formatBytes(int64(bytesPerSec)) + "/s"
}

// formatTime renders a duration in seconds as a short human string.
func formatTime(seconds *int) string {
	if seconds == nil || *seconds < 0 {
		return "Unknown"
	}
	if *seconds == 0 {
		return "Done"
	}

	rest := *seconds
	days := rest / (3600 * 24)
	rest -= days * 3600 * 24
	hours := rest / 3600
	rest -= hours * 3600
	minutes := rest / 60
	secs := rest - minutes*60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if secs > 0 {
		parts = append(parts, fmt.Sprintf("%ds", secs))
	}
	if len(parts) == 0 {
		return "Unknown"
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}

func formatProgress(p float64) string {
	return fmt.Sprintf("%.1f", p)
}

func fmtInt(n int) string {
	return strconv.Itoa(n)
}

func progressPercent(done, total int64) float64 {
	if total <= 0 {
		return 0
	}
	p := float64(done) / float64(total) * 100
	if p < 0 {
		return 0
	}
	return min(p, 100)
}
