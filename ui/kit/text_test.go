package kit

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/task"
)

func TestFormatUptime(t *testing.T) {
	t.Cleanup(func() { SetLang("en") })

	SetLang("zh")
	if got := FormatUptime(41*time.Hour + 50*time.Minute); got != "1天 17小时 50分钟" {
		t.Fatalf("zh: %q", got)
	}
	if got := FormatUptime(0); got != "-" {
		t.Fatalf("zero: %q", got)
	}

	SetLang("en")
	if got := FormatUptime(41*time.Hour + 50*time.Minute); got != "1d 17h 50m" {
		t.Fatalf("en: %q", got)
	}
	if got := FormatUptime(5 * time.Minute); got != "0d 00h 05m" {
		t.Fatalf("5m: %q", got)
	}
}

func TestPadAndBytes(t *testing.T) {
	if w := lipgloss.Width(PadRight("密码", 10)); w != 10 {
		t.Fatalf("pad width %d", w)
	}
	if got := HumanBytes(2048); got != "2.0K" {
		t.Fatalf("human: %q", got)
	}
}

func TestGeoCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache")
	c := task.Open(path)
	if err := c.Put("geo", "jp"); err != nil {
		t.Fatal(err)
	}
	old := geoCache
	geoCache = func() *task.Cache { return c }
	t.Cleanup(func() { geoCache = old })

	// A cached answer is returned without touching the network.
	code, ok := geoCountry()
	if !ok || code != "JP" {
		t.Fatalf("geo: %q %v", code, ok)
	}
}
