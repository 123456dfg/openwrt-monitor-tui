package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/language"
	"openwrt-monitor-tui/ui/settings/startup"
	"openwrt-monitor-tui/ui/settings/system"
	"openwrt-monitor-tui/ui/settings/wifi"
)

// mouse.go maps pointer positions onto the same actions the keyboard uses.
// The x/y ranges come from the view packages, so a layout change has to update
// one place only.
func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.scroll = m.clampScroll(m.scroll - mouseScroll)
	case tea.MouseButtonWheelDown:
		m.scroll = m.clampScroll(m.scroll + mouseScroll)
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			break
		}
		m.commitEdit()
		switched := false
		switch {
		case msg.Y == tabRowY:
			for i, h := range m.tabHits() {
				if msg.X >= h[0] && msg.X <= h[1] {
					m.setTab(i)
					switched = true
					break
				}
			}
		case m.tab == tabSettings:
			row := msg.Y - bandRowY + m.scroll
			if row < 0 || row >= m.rows() {
				break
			}
			if msg.X < kit.SidebarW+2 {
				// Mouse click selects and opens in one go; only the keyboard
				// needs Enter to confirm the highlighted row.
				m.clickSubtab(row)
			} else {
				// Clicking inside the pane implies entering it, even after Esc
				// returned the focus to the sidebar.
				m.subActive = true
				switch m.subOpen {
				case settings.Language:
					m.clickLang(msg.X, row)
				case settings.Cron:
					if cmd := m.clickCron(msg.X, row); cmd != nil {
						return m, cmd
					}
				case settings.Startup:
					if cmd := m.clickStartup(msg.X, row); cmd != nil {
						return m, cmd
					}
				case settings.System:
					if cmd := m.clickSystem(msg.X, row); cmd != nil {
						return m, cmd
					}
				default: // Wi-Fi
					if row == 0 {
						for i, h := range wifi.BandHits(m.snap.Wifi.Bands, m.band) {
							if msg.X >= h[0] && msg.X <= h[1] {
								m.selectBand(i)
								break
							}
						}
						break
					}
					if cmd := m.clickSettings(msg.X, row); cmd != nil {
						return m, cmd
					}
				}
			}
		}
		if switched {
			return m, m.refreshSoon()
		}
	}
	return m, nil
}

// clickSettings runs a Wi-Fi form row under the pointer. Field rows only take
// focus; only the Apply/Cancel rows act.
func (m *model) clickSettings(x, row int) tea.Cmd {
	if x < kit.SidebarW+2 || m.form.Busy || len(m.snap.Wifi.Bands) == 0 {
		return nil
	}
	off := wifi.FormRowOffset(m.settingsState().Wifi)
	switch {
	case row == off+wifi.FocusEnabled:
		// The switch is a checkbox: a click flips it right away.
		m.form.Focus = wifi.FocusEnabled
		mm, cmd := m.activate()
		*m = mm.(model)
		return cmd
	case row > off+wifi.FocusEnabled && row <= off+wifi.FocusSecurity:
		m.form.Focus = row - off
		// A click on a text input is the whole gesture: focus it and start
		// editing right away, without changing the value. The cycling rows
		// still only take focus.
		if m.form.Focus == wifi.FocusSSID || m.form.Focus == wifi.FocusPassword {
			if m.form.Iface == "" {
				m.formNote = kit.Tr.NoAP
				return nil
			}
			if in := m.form.Text(); in != nil {
				in.Begin()
			}
			m.form.Editing = true
		}
		return nil
	case row == off+wifi.FocusApply+1:
		// Apply and Cancel share one row; pick by x.
		ax := x - (kit.SidebarW + 2)
		apply, cancel := wifi.ButtonHits()
		switch {
		case ax >= apply[0] && ax <= apply[1]:
			m.form.Focus = wifi.FocusApply
		case ax >= cancel[0] && ax <= cancel[1]:
			m.form.Focus = wifi.FocusCancel
		default:
			return nil
		}
	default:
		return nil
	}
	mm, cmd := m.activate()
	*m = mm.(model)
	return cmd
}

// clickCron maps a pointer position onto the sub-tab strip, the job list or
// the form. Text rows start editing in the same click; pickers only take focus.
func (m *model) clickCron(x, row int) tea.Cmd {
	if x < kit.SidebarW+2 {
		return nil
	}
	st := m.settingsState().Cron
	if row == 0 {
		for i, h := range cron.TabHits(st) {
			if x >= h[0] && x <= h[1] {
				m.cronTab = i
				m.cronNote, m.cronErr = "", ""
				break
			}
		}
		return nil
	}
	if m.cronTab == cron.TabList {
		i := row - cron.ContentOffset(st)
		if !st.Avail || i < 0 || i >= len(st.Jobs) {
			return nil
		}
		m.cronCursor = i
		ax := x - (kit.SidebarW + 2)
		check := cron.CheckHits()
		if ax >= check[0] && ax <= check[1] {
			return m.cronToggle(st.Jobs[i])
		}
		m.cronLoad()
		return nil
	}
	if m.cron.Busy {
		return nil
	}
	off := cron.FormRowOffset(st)
	if i := row - off; i >= 0 && i < len(m.cron.Fields()) {
		m.cron.Focus = m.cron.Fields()[i]
		if in := m.cron.Text(); in != nil {
			in.Begin()
			m.cron.Editing = true
		}
		return nil
	}
	if row == cron.ButtonsRow(st) {
		ax := x - (kit.SidebarW + 2)
		apply, cancel := m.cron.Buttons().Hits()
		switch {
		case ax >= apply[0] && ax <= apply[1]:
			m.cron.Focus = cron.FocusApply
		case ax >= cancel[0] && ax <= cancel[1]:
			m.cron.Focus = cron.FocusCancel
		default:
			return nil
		}
		mm, cmd := m.activateCron()
		*m = mm.(model)
		return cmd
	}
	return nil
}

// clickStartup maps a pointer position onto the line editor. A line starts
// editing in the same click; the buttons run on click.
func (m *model) clickStartup(x, row int) tea.Cmd {
	f := &m.startup
	if !f.Available || f.Busy {
		return nil
	}
	st := m.settingsState().Startup
	off := startup.LineOffset(st)
	n := f.TaskCount()

	if row >= off && row < off+n {
		f.Focus = row - off
		ax := x - (kit.SidebarW + 2)
		check := startup.CheckHits()
		if ax >= check[0] && ax <= check[1] {
			f.Toggle(f.Focus)
			return nil
		}
		f.Begin(f.Focus)
		return nil
	}
	if row == off+max(1, n)+1 {
		ax := x - (kit.SidebarW + 2)
		apply, cancel := startup.ButtonHits()
		switch {
		case ax >= apply[0] && ax <= apply[1]:
			f.Focus = f.ApplyIndex()
		case ax >= cancel[0] && ax <= cancel[1]:
			f.Focus = f.CancelIndex()
		default:
			return nil
		}
		mm, cmd := m.activateStartup()
		*m = mm.(model)
		return cmd
	}
	return nil
}

// clickLang selects a row of the language picker; clicking a radio is the
// expected way to choose.
func (m *model) clickLang(x, row int) {
	if x < kit.SidebarW+2 {
		return
	}
	if i := row - 1; i >= 0 && i < len(language.Choices) {
		m.chooseLang(i)
	}
}

// clickSubtab selects and enters a sidebar row in one step: a mouse click is
// itself the confirmation, unlike the keyboard which needs Enter.
func (m *model) clickSubtab(i int) {
	secs := m.sections()
	if i < 0 || i >= len(secs) {
		return
	}
	want := secs[i]
	m.selectSubtab(i) // may refuse to leave a dirty Wi-Fi form
	if want != m.subOpen {
		return
	}
	m.subSel = i
	m.openSubtab()
}

// clickSystem selects an action; clicking the selected one (or the
// confirmation line) runs it. That mirrors the two Enters of the keyboard.
func (m *model) clickSystem(x, row int) tea.Cmd {
	if x < kit.SidebarW+2 || m.sysBusy {
		return nil
	}
	if i := row - system.LineOffset(); i >= 0 && i < int(system.Count) {
		if i == m.sysSel && m.sysConfirm {
			return m.runSelectedSystemAction()
		}
		m.sysSel = i
		m.sysConfirm = true
		m.sysNote, m.sysErr = "", ""
		return nil
	}
	if row == system.ConfirmRow() && m.sysConfirm {
		return m.runSelectedSystemAction()
	}
	return nil
}

func (m *model) runSelectedSystemAction() tea.Cmd {
	mm, cmd := m.updateSystem(tea.KeyMsg{Type: tea.KeyEnter})
	*m = mm.(model)
	return cmd
}
