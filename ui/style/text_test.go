package style

import (
	"openwrt-monitor-tui/ui/i18n"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestFormatUptime(t *testing.T) {
	t.Cleanup(func() { i18n.SetLang("en") })

	i18n.SetLang("zh")
	if got := FormatUptime(41*time.Hour + 50*time.Minute); got != "1天 17小时 50分钟" {
		t.Fatalf("zh: %q", got)
	}
	if got := FormatUptime(0); got != "-" {
		t.Fatalf("zero: %q", got)
	}

	i18n.SetLang("en")
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
