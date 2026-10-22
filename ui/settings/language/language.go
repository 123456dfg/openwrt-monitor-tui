package language

import (
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/widgets"
)

// Choices are the values behind the picker rows, in display order.
var Choices = []string{"auto", "en", "zh"}

func Options() []string {
	return []string{
		kit.Tr.Auto + " (" + langName(kit.LangAutoCode()) + ")",
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
	r := widgets.NewRadio(sel)
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
