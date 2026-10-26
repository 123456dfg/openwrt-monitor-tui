package ui

import (
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/network"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

func (m model) View() string {
	if m.w <= 0 {
		return style.Tr.Loading
	}

	head := " " + style.Accent.Render("openwrt-monitor-tui") + style.Dim.Render("  "+style.Glyph.Dot+"  ") +
		style.Fallback(m.snap.Release, "OpenWrt") + style.Dim.Render("  "+style.Glyph.Dot+"  ") +
		style.Fallback(m.snap.Arch, runtime.GOARCH) +
		style.Dim.Render("  "+style.Glyph.Dot+"  Linux "+style.Fallback(m.snap.Kernel, "?"))
	head = lipgloss.NewStyle().MaxWidth(m.w).Render(head)

	tabs := m.topBar()

	lines := m.body()
	off := min(m.scroll, len(lines))
	visible := lines[off:]
	if len(visible) > m.rows() {
		visible = visible[:m.rows()]
	}
	for i, ln := range visible {
		if ln != "" {
			visible[i] = fit(ln, m.w)
		}
	}
	for len(visible) < m.rows() {
		visible = append(visible, "")
	}

	hint := style.Tr.HintScroll
	if m.tab == 1 {
		switch {
		case !m.netActive:
			hint = style.Tr.HintSidebar
		case m.netOpen == network.TabSpeed:
			hint = style.Tr.HintSpeed
		case m.netOpen == network.TabPorts:
			hint = style.Tr.HintPorts
		}
	}
	if m.tab == 2 {
		hint = style.Tr.HintProcs
	}
	if m.tab == tabSettings {
		switch {
		case !m.subActive:
			hint = style.Tr.HintSidebar
		case m.subOpen == settings.Language:
			hint = style.Tr.HintLang
		case m.subOpen == settings.Cron:
			hint = style.Tr.HintCron
			if m.cronTab == cron.TabList {
				hint = style.Tr.HintCronList
			}
		case m.subOpen == settings.Startup:
			hint = style.Tr.HintStartup
		case m.subOpen == settings.System:
			hint = style.Tr.HintSystem
		default:
			hint = style.Tr.HintSettings
		}
	}
	right := style.Tr.Uptime + " " + style.FormatUptime(m.snap.Uptime) + "  " +
		style.Tr.RefreshInterval + " " + style.FormatMS(m.snap.Took)
	status := " " + style.Dim.Render(hint)
	if m.w > 60 && lipgloss.Width(right)+20 < m.w {
		// Pin the timing block to the right edge and give the hint whatever
		// room is left.
		avail := m.w - lipgloss.Width(right) - 3
		left := truncToWidth(hint, avail)
		pad := m.w - 1 - lipgloss.Width(left) - lipgloss.Width(right)
		status = " " + style.Dim.Render(left+strings.Repeat(" ", pad)+right)
	}
	status = fit(status, m.w)

	return strings.Join([]string{
		head,
		style.Dim.Render(strings.Repeat(style.Glyph.Sep, m.w)),
		tabs,
		style.Dim.Render(strings.Repeat(style.Glyph.Sep, m.w)),
		strings.Join(visible, "\n"),
		style.Dim.Render(strings.Repeat(style.Glyph.Sep, m.w)),
		status,
	}, "\n")
}

// topBar is the tab row with the startup cost pinned to the right edge.
func (m model) topBar() string {
	bar := " " + m.tabBar()
	if m.boot <= 0 {
		return bar
	}
	right := style.Tr.Boot + " " + style.FormatMS(m.boot)
	used := lipgloss.Width(bar)
	if m.w <= used+lipgloss.Width(right)+2 {
		return bar
	}
	pad := m.w - used - lipgloss.Width(right)
	return bar + strings.Repeat(" ", pad) + style.Dim.Render(right)
}

func (m model) tabBar() string {
	return components.Tabs{Names: tabNames(), Cur: m.tab}.View()
}

// fit truncates a styled line to w cells without cutting escape sequences.
func fit(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	return lipgloss.NewStyle().MaxWidth(w).Render(s)
}

// truncToWidth cuts plain text to at most w terminal cells, so a CJK hint does
// not overflow when the right-aligned status block needs room.
func truncToWidth(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	w-- // room for the ellipsis
	if w <= 0 {
		return ""
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		rw := lipgloss.Width(string(r))
		if used+rw > w {
			break
		}
		b.WriteRune(r)
		used += rw
	}
	return b.String() + "…"
}
