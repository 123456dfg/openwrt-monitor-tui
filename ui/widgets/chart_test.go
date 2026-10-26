package widgets

import (
	"openwrt-monitor-tui/ui/i18n"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/style"
)

func TestChartBraille(t *testing.T) {
	i18n.SetLang("en")
	style.SetUnicode()
	t.Cleanup(style.SetUnicode)

	values := []float64{0, 10, 25, 60, 30, 5, 0}
	c := Chart(values, 12)
	if lipgloss.Width(c) != 12 {
		t.Fatalf("width %d", lipgloss.Width(c))
	}
	for _, r := range c {
		if r < 0x2800 || r > 0x28FF {
			t.Fatalf("not braille: %U", r)
		}
	}

	// an empty series is a flat baseline of the right width
	empty := Chart(nil, 12)
	if lipgloss.Width(empty) != 12 {
		t.Fatalf("empty width %d", lipgloss.Width(empty))
	}
	if empty != strings.Repeat("⣀", 12) {
		t.Fatalf("empty series is not a baseline: %q", empty)
	}

	// more samples than cells keeps the newest ones
	long := make([]float64, 40)
	for i := range long {
		long[i] = float64(i)
	}
	if got := Chart(long, 12); lipgloss.Width(got) != 12 {
		t.Fatalf("long series width %d", lipgloss.Width(got))
	}
}

func TestChartASCII(t *testing.T) {
	style.SetASCII()
	t.Cleanup(style.SetUnicode)

	c := Chart([]float64{1, 2, 3}, 6)
	if strings.ContainsRune(c, 0x2800) {
		t.Fatalf("braille in ASCII mode: %q", c)
	}
	if lipgloss.Width(c) != 6 {
		t.Fatalf("width %d", lipgloss.Width(c))
	}
	if got := Spark([]float64{0, 1, 2}, 3); got == "" {
		t.Fatal("empty spark")
	}
}
