package ui

import (
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
)

func (m model) View() string {
	if m.w <= 0 {
		return kit.Tr.Loading
	}

	head := " " + kit.Accent.Render("openwrt-monitor-tui") + kit.Dim.Render("  "+kit.Glyph.Dot+"  ") +
		kit.Fallback(m.snap.Release, "OpenWrt") + kit.Dim.Render("  "+kit.Glyph.Dot+"  ") +
		kit.Fallback(m.snap.Arch, runtime.GOARCH) +
		kit.Dim.Render("  "+kit.Glyph.Dot+"  Linux "+kit.Fallback(m.snap.Kernel, "?"))
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

	hint := kit.Tr.HintScroll
	if m.tab == tabSettings {
		switch {
		case !m.subActive:
			hint = kit.Tr.HintSidebar
		case m.subOpen == settings.Language:
			hint = kit.Tr.HintLang
		case m.subOpen == settings.Cron:
			hint = kit.Tr.HintCron
			if m.cronTab == cron.TabList {
				hint = kit.Tr.HintCronList
			}
		case m.subOpen == settings.Startup:
			hint = kit.Tr.HintStartup
		case m.subOpen == settings.System:
			hint = kit.Tr.HintSystem
		default:
			hint = kit.Tr.HintSettings
		}
	}
	right := kit.Tr.Uptime + " " + kit.FormatUptime(m.snap.Uptime) + "  " +
		kit.Tr.RefreshInterval + " " + kit.FormatMS(m.snap.Took)
	status := " " + kit.Dim.Render(hint)
	if m.w > 60 && lipgloss.Width(right)+20 < m.w {
		// Pin the timing block to the right edge and give the hint whatever
		// room is left.
		avail := m.w - lipgloss.Width(right) - 3
		left := truncToWidth(hint, avail)
		pad := m.w - 1 - lipgloss.Width(left) - lipgloss.Width(right)
		status = " " + kit.Dim.Render(left+strings.Repeat(" ", pad)+right)
	}
	status = fit(status, m.w)

	return strings.Join([]string{
		head,
		kit.Dim.Render(strings.Repeat(kit.Glyph.Sep, m.w)),
		tabs,
		kit.Dim.Render(strings.Repeat(kit.Glyph.Sep, m.w)),
		strings.Join(visible, "\n"),
		kit.Dim.Render(strings.Repeat(kit.Glyph.Sep, m.w)),
		status,
	}, "\n")
}

// topBar is the tab row with the startup cost pinned to the right edge.
func (m model) topBar() string {
	bar := " " + m.tabBar()
	if m.boot <= 0 {
		return bar
	}
	right := kit.Tr.Boot + " " + kit.FormatMS(m.boot)
	used := lipgloss.Width(bar)
	if m.w <= used+lipgloss.Width(right)+2 {
		return bar
	}
	pad := m.w - used - lipgloss.Width(right)
	return bar + strings.Repeat(" ", pad) + kit.Dim.Render(right)
}

func (m model) tabBar() string {
	var b strings.Builder
	for i, n := range tabNames() {
		if i > 0 {
			b.WriteString("   ")
		}
		if i == m.tab {
			if kit.ColorOn() {
				b.WriteString(kit.Accent.Render(n))
			} else {
				b.WriteString("[" + n + "]")
			}
			continue
		}
		b.WriteString(kit.Label.Render(n))
	}
	return b.String()
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
