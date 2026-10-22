package settings

import (
	"strings"

	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/language"
	"openwrt-monitor-tui/ui/settings/startup"
	"openwrt-monitor-tui/ui/settings/system"
	"openwrt-monitor-tui/ui/settings/wifi"
)

type Section int

const (
	Wifi Section = iota
	Cron
	Startup
	System
	Language
)

// Sections is the sidebar. Wi-Fi disappears on devices without a radio; the
// crontab and startup pages are file based, so they are always listed.
func Sections(wifiSupported bool) []Section {
	var out []Section
	if wifiSupported {
		out = append(out, Wifi)
	}
	return append(out, Cron, Startup, System, Language)
}

func Name(s Section) string {
	switch s {
	case Cron:
		return kit.Tr.Cron
	case Startup:
		return kit.Tr.Startup
	case System:
		return kit.Tr.System
	case Language:
		return kit.Tr.Language
	}
	return "Wi-Fi"
}

// State is everything the settings tab needs to render one frame.
type State struct {
	Sections  []Section
	Sel       int // sidebar cursor
	Open      Section
	Active    bool // content has the focus
	Wifi      wifi.State
	Cron      cron.State
	Startup   startup.State
	System    system.State
	LangSel   int
	LangFocus int
}

// Body merges the sidebar and the pane of the open section.
func Body(st State) []string {
	subs := st.Sections
	if len(subs) == 0 {
		subs = Sections(false)
	}
	cur := min(st.Sel, len(subs)-1)

	side := make([]string, 0, len(subs))
	for i, sec := range subs {
		if i == cur {
			side = append(side, " "+kit.Accent.Render(kit.Glyph.Sel+" "+Name(sec)))
			continue
		}
		side = append(side, "  "+kit.Label.Render(Name(sec)))
	}

	var content []string
	switch st.Open {
	case Wifi:
		content = wifi.Pane(st.Wifi)
	case Cron:
		content = cron.Pane(st.Cron)
	case Startup:
		content = startup.Pane(st.Startup)
	case System:
		content = system.Pane(st.System)
	default:
		content = language.Pane(st.LangSel, st.LangFocus, st.Active)
	}

	out := make([]string, max(len(side), len(content)))
	for i := range out {
		var left, right string
		if i < len(side) {
			left = side[i]
		}
		if i < len(content) {
			right = content[i]
		}
		out[i] = strings.TrimRight(kit.PadRight(left, kit.SidebarW)+"  "+right, " ")
	}
	return out
}
