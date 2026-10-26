package language

import (
	"openwrt-monitor-tui/ui/i18n"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// Choices are the values behind the picker rows, in display order.
var Choices = []string{"auto", "en", "zh"}

func Options() []string {
	return []string{
		style.Tr.Auto + " (" + langName(i18n.LangAutoCode()) + ")",
		"English",
		"中文",
	}
}

func langName(code string) string {
	if code == "zh" {
		return "中文"
	}
	return "English"
}

// Pane renders the language picker. The active pick is marked with (o); the
// cursor only appears once the content has the focus.
func Pane(sel, focus int, active bool) []string {
	r := components.NewRadio(sel)
	if active {
		r.Cursor = focus
	} else {
		r.Cursor = -1
	}
	out := []string{""}
	for _, line := range r.Lines(Options()) {
		out = append(out, "  "+line)
	}
	return out
}
