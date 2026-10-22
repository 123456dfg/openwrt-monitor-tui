package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/language"
	"openwrt-monitor-tui/ui/settings/wifi"
)

// handleKey is the keyboard side of the UI. It moves a focus cursor; the mouse
// side resolves clicks from the rendered coordinates instead. Keeping the two
// apart is what stops a hit box from silently disagreeing with a binding.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.tab == tabSettings && m.editing() {
		return m.updateEdit(msg)
	}
	if msg.String() == "esc" && m.tab == tabSettings && m.subActive {
		// On the system page Esc first drops the pending confirmation.
		if m.subOpen == settings.System {
			if m.sysConfirm {
				m.sysConfirm = false
				return m, nil
			}
			if m.sysBusy {
				return m, nil
			}
		}
		m.subActive = false
		return m, nil
	}
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m, m.collectCmd(false)
	case "1", "2", "3", "4":
		m.setTab(int(msg.String()[0] - '1'))
		return m, m.refreshSoon()
	case "tab":
		m.setTab((m.tab + 1) % len(tabNames()))
		return m, m.refreshSoon()
	case "shift+tab":
		m.setTab((m.tab + len(tabNames()) - 1) % len(tabNames()))
		return m, m.refreshSoon()
	}

	if m.tab == tabSettings {
		// The sidebar is still part of the main tab: arrows keep their usual
		// meaning there, so the band selector only takes them over once the
		// Wi-Fi section itself has been entered.
		if !m.subActive {
			switch msg.String() {
			case "up", "k":
				m.moveSubtab(-1)
			case "down", "j":
				m.moveSubtab(+1)
			case "left", "h":
				return m.switchTab(-1)
			case "right", "l":
				return m.switchTab(+1)
			case "enter", " ":
				m.openSubtab()
			}
			return m, nil
		}
		switch m.subOpen {
		case settings.Language:
			switch msg.String() {
			case "up", "k":
				m.langFocus = wrapIndex(m.langFocus, -1, len(language.Choices))
			case "down", "j":
				m.langFocus = wrapIndex(m.langFocus, +1, len(language.Choices))
			case "left", "h":
				return m.switchTab(-1)
			case "right", "l":
				return m.switchTab(+1)
			case "enter", " ":
				m.chooseLang(m.langFocus)
			}
			return m, nil
		case settings.Cron:
			return m.updateCron(msg)
		case settings.System:
			return m.updateSystem(msg)
		case settings.Startup:
			switch msg.String() {
			case "left", "h":
				return m.switchTab(-1)
			case "right", "l":
				return m.switchTab(+1)
			}
			return m.updateStartup(msg)
		}
		switch msg.String() {
		case "up", "k":
			m.form.Focus = wrapIndex(m.form.Focus, -1, wifi.FocusMax)
		case "down", "j":
			m.form.Focus = wrapIndex(m.form.Focus, +1, wifi.FocusMax)
		case "left", "h":
			if !m.toggleButtonFocus() {
				m.stepBand(-1)
			}
		case "right", "l":
			if !m.toggleButtonFocus() {
				m.stepBand(+1)
			}
		case "enter", " ":
			return m.activate()
		case "p":
			m.showKeys = !m.showKeys
		}
		return m, nil
	}

	switched := false
	switch msg.String() {
	case "l", "right":
		m.setTab((m.tab + 1) % len(tabNames()))
		switched = true
	case "h", "left":
		m.setTab((m.tab + len(tabNames()) - 1) % len(tabNames()))
		switched = true
	case "down", "j":
		m.scroll = m.clampScroll(m.scroll + 1)
	case "up", "k":
		m.scroll = m.clampScroll(m.scroll - 1)
	case "pgdown", " ":
		m.scroll = m.clampScroll(m.scroll + m.rows())
	case "pgup":
		m.scroll = m.clampScroll(m.scroll - m.rows())
	case "g":
		m.scroll = 0
	case "G":
		m.scroll = m.clampScroll(1 << 20)
	}
	if switched {
		return m, m.refreshSoon()
	}
	return m, nil
}

// switchTab moves to the previous/next main tab and samples the new one.
func (m model) switchTab(d int) (tea.Model, tea.Cmd) {
	m.setTab((m.tab + d + len(tabNames())) % len(tabNames()))
	return m, m.refreshSoon()
}
