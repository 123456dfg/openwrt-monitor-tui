package kit

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// PadRight and PadLeft pad to terminal cells, so CJK labels line up with ASCII
// ones. fmt's %-10s counts runes, not cells, and wide characters would shift
// every column.
func PadRight(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return s + strings.Repeat(" ", w-n)
	}
	return s
}

func PadLeft(s string, w int) string {
	if n := lipgloss.Width(s); n < w {
		return strings.Repeat(" ", w-n) + s
	}
	return s
}

func HumanBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%dB", v)
	}
	f := float64(v)
	for _, u := range []string{"K", "M", "G", "T"} {
		f /= unit
		if f < unit {
			return fmt.Sprintf("%.1f%s", f, u)
		}
	}
	return fmt.Sprintf("%.1fP", f/unit)
}

func Fallback(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

// FormatUptime renders a duration as dd天 hh小时 mm分钟 (or the English
// equivalent). Seconds are deliberately omitted.
func FormatUptime(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	days := int(d.Hours()) / 24
	h := int(d.Hours()) % 24
	m := int(d.Minutes()) % 60
	return fmt.Sprintf(Tr.UptimeFmt, days, h, m)
}

// FormatMS renders a sample duration in whole milliseconds.
func FormatMS(d time.Duration) string {
	return fmt.Sprintf("%d ms", d.Milliseconds())
}
