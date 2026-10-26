package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/ui/network"
	"openwrt-monitor-tui/ui/network/portscan"
	"openwrt-monitor-tui/ui/network/speedtest"
	"openwrt-monitor-tui/ui/procs"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/language"
	"openwrt-monitor-tui/ui/settings/startup"
	"openwrt-monitor-tui/ui/settings/system"
	"openwrt-monitor-tui/ui/settings/wifi"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// mouse.go maps pointer positions onto the same actions the keyboard uses.
// The components own their hit-testing and state changes (Clickable /
// Wheelable in ui/widgets/components); this file only translates screen
// coordinates into component coordinates and orchestrates what a state change
// means — a command, a note, a focus move.
func (m model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		// The wheel switches any option box under the pointer; elsewhere it
		// scrolls the page.
		if !m.wheelStep(msg, -1) {
			m.scroll = m.clampScroll(m.scroll - mouseScroll)
		}
	case tea.MouseButtonWheelDown:
		if !m.wheelStep(msg, +1) {
			m.scroll = m.clampScroll(m.scroll + mouseScroll)
		}
	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			break
		}
		m.commitEdit()
		switched := false
		switch {
		case msg.Y == tabRowY:
			// The main tab strip is a components.Tabs.
			if i := (components.Tabs{Names: tabNames(), Cur: m.tab}).Click(1, msg.X); i >= 0 {
				m.setTab(i)
				switched = true
			}
		case m.tab == 1:
			row := msg.Y - bandRowY + m.scroll
			if row < 0 || row >= m.rows() {
				break
			}
			if msg.X < style.SidebarW+2 {
				// Sidebar: a click selects and enters in one go, matching the
				// settings tab.
				if i := row; i >= network.TabOverview && i <= network.TabPorts {
					m.netSel, m.netOpen, m.netActive = i, i, true
				}
				break
			}
			// Clicking inside the pane implies entering it, even after Esc
			// returned the focus to the sidebar.
			m.netActive = true
			if m.netOpen == network.TabSpeed {
				if cmd := m.clickSpeed(row); cmd != nil {
					return m, cmd
				}
			}
			if m.netOpen == network.TabPorts {
				if cmd := m.clickPorts(row); cmd != nil {
					return m, cmd
				}
			}
		case m.tab == 2:
			// The process table header sorts: click a column to rank by it,
			// click the active column again to flip the direction.
			if row := msg.Y - bandRowY + m.scroll; row == procs.HeaderRow {
				if by, ok := procs.HeaderClick(msg.X); ok {
					m.sortProcs(by)
				}
			}
		case m.tab == tabSettings:
			row := msg.Y - bandRowY + m.scroll
			if row < 0 || row >= m.rows() {
				break
			}
			if msg.X < style.SidebarW+2 {
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
// focus; only the Apply/Cancel rows act. The form owns the widgets, so the
// events go straight to them.
func (m *model) clickSettings(x, row int) tea.Cmd {
	if x < style.SidebarW+2 || m.form.Busy || len(m.snap.Wifi.Bands) == 0 {
		return nil
	}
	off := wifi.FormRowOffset(m.settingsState().Wifi)
	switch {
	case row == off+wifi.FocusEnabled:
		// The switch is a checkbox: a click flips it right away.
		m.form.Focus = wifi.FocusEnabled
		m.form.Enabled.Click(x, row)
		m.formNote, m.formErr = "", ""
		return nil
	case row > off+wifi.FocusEnabled && row <= off+wifi.FocusSecurity:
		m.form.Focus = row - off
		// A click on a text input is the whole gesture: the component starts
		// the editing session, without changing the value. The cycling rows
		// still only take focus.
		if m.form.Focus == wifi.FocusSSID || m.form.Focus == wifi.FocusPassword {
			if m.form.Iface == "" {
				m.formNote = style.Tr.NoAP
				return nil
			}
			if in := m.form.Text(); in != nil {
				in.Click(x, row)
				m.form.Editing = true
			}
		}
		return nil
	case row == off+wifi.FocusApply+1:
		// Apply and Cancel share one row; the Buttons component picks.
		ax := x - (style.SidebarW + 2)
		btns := components.Buttons{Labels: [2]string{style.Tr.Apply, style.Tr.Cancel}}
		if !btns.Click(ax, 0) {
			return nil
		}
		if btns.Focus == 0 {
			m.form.Focus = wifi.FocusApply
		} else {
			m.form.Focus = wifi.FocusCancel
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
	if x < style.SidebarW+2 {
		return nil
	}
	st := m.settingsState().Cron
	if row == 0 {
		// The strip is a components.Tabs.
		if i := cron.TabClick(st, x); i >= 0 {
			m.cronTab = i
			m.cronNote, m.cronErr = "", ""
		}
		return nil
	}
	if m.cronTab == cron.TabList {
		i := row - cron.ContentOffset(st)
		if !st.Avail || i < 0 || i >= len(st.Jobs) {
			return nil
		}
		m.cronCursor = i
		// The job switch is a checkbox drawn at columns 4..6; the component
		// resolves the press, the pane persists the flip in the background.
		ax := x - (style.SidebarW + 2)
		if ax >= 4 && ax <= 6 {
			cb := components.Checkbox{On: st.Jobs[i].Enabled}
			cb.Click(ax-4, row)
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
			in.Click(x, row)
			m.cron.Editing = true
		}
		return nil
	}
	if row == cron.ButtonsRow(st) {
		ax := x - (style.SidebarW + 2)
		btns := m.cron.Buttons()
		if !btns.Click(ax, 0) {
			return nil
		}
		if btns.Focus == 0 {
			m.cron.Focus = cron.FocusApply
		} else {
			m.cron.Focus = cron.FocusCancel
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
		// The task switch is a checkbox at columns 4..6; the component
		// resolves the press, the form persists the flip.
		ax := x - (style.SidebarW + 2)
		if ax >= 4 && ax <= 6 {
			cb := components.Checkbox{On: true}
			cb.Click(ax-4, row)
			f.Toggle(f.Focus)
			return nil
		}
		f.Begin(f.Focus)
		return nil
	}
	if row == off+max(1, n)+1 {
		ax := x - (style.SidebarW + 2)
		btns := f.Buttons()
		if !btns.Click(ax, 0) {
			return nil
		}
		if btns.Focus == 0 {
			f.Focus = f.ApplyIndex()
		} else {
			f.Focus = f.CancelIndex()
		}
		mm, cmd := m.activateStartup()
		*m = mm.(model)
		return cmd
	}
	return nil
}

// clickLang selects a row of the language picker through the radio component;
// clicking a radio is the expected way to choose.
func (m *model) clickLang(x, row int) {
	if x < style.SidebarW+2 {
		return
	}
	r := components.Radio{Options: language.Choices, Cursor: m.langFocus, Sel: m.langSel}
	if !r.Click(x, row-1) {
		return
	}
	m.langFocus = r.Cursor
	m.chooseLang(r.Sel)
}

// wheelStep switches the option widget under the pointer with the wheel —
// d is -1 for wheel-up, +1 for wheel-down. It reports whether the pointer was
// over an option box; false lets the wheel scroll the page as before.
func (m *model) wheelStep(msg tea.MouseMsg, d int) bool {
	row := msg.Y - bandRowY + m.scroll
	switch {
	case m.tab == tabSettings && !m.subActive:
		// The sidebar is one long option list; the wheel walks it.
		if msg.X >= style.SidebarW+2 {
			return false
		}
		m.moveSubtab(d)
		return true

	case m.tab == 1 && !m.netActive:
		// The network sidebar is an option list too.
		if msg.X >= style.SidebarW+2 {
			return false
		}
		m.moveNetSel(d)
		return true

	case m.tab == tabSettings && m.subActive:
		switch m.subOpen {
		case settings.Language:
			// The wheel moves the radio cursor; Enter still applies.
			if i := row - 1; i >= 0 && i < len(language.Choices) {
				r := components.Radio{Options: language.Choices, Cursor: m.langFocus, Sel: m.langSel}
				r.Wheel(d)
				m.langFocus, m.langSel = r.Cursor, r.Sel
				return true
			}
		case settings.Cron:
			return m.wheelCron(row, d)
		case settings.System:
			if i := row - system.LineOffset(); i >= 0 && i < int(system.Count) {
				m.sysSel = wrapIndex(m.sysSel, d, int(system.Count))
				m.sysConfirm = false
				return true
			}
		default: // Wi-Fi
			return m.wheelWifi(row, d)
		}
	}
	return false
}

// wheelCron switches the cron option boxes: the job list cursor on the list
// tab, the schedule-kind and weekday pickers in the editor. Values only move
// once the row has been entered, matching the keyboard rule.
func (m *model) wheelCron(row, d int) bool {
	st := m.settingsState().Cron
	if m.cronTab == cron.TabList {
		i := row - cron.ContentOffset(st)
		if i < 0 || i >= len(st.Jobs) {
			return false
		}
		m.cronCursor = wrapIndex(m.cronCursor, d, len(st.Jobs))
		return true
	}
	if m.cron.Busy {
		return false
	}
	if i := row - cron.FormRowOffset(st); i >= 0 && i < len(m.cron.Fields()) {
		field := m.cron.Fields()[i]
		if field != cron.FocusKind && field != cron.FocusWeekday {
			return false // text rows are not option boxes
		}
		m.cron.Focus = field
		if m.cron.Editing {
			if sel := m.cron.Picker(); sel != nil {
				sel.Wheel(d)
			}
		}
		return true
	}
	return false
}

// wheelWifi switches the channel and security pickers; like the keyboard,
// values move only while the row is entered, otherwise the wheel just moves
// the focus there.
func (m *model) wheelWifi(row, d int) bool {
	off := wifi.FormRowOffset(m.settingsState().Wifi)
	i := row - off
	if i != wifi.FocusChannel && i != wifi.FocusSecurity {
		return false
	}
	m.form.Focus = i
	if m.form.Editing {
		if sel := m.form.Picker(); sel != nil {
			sel.Wheel(d)
		}
	}
	return true
}

// clickSpeed maps a pointer position onto the speed test pane: the whole
// action row is the hit area, in either visual form (button or play-style).
// The pane starts at body row 0, aligned with the sidebar.
func (m *model) clickSpeed(row int) tea.Cmd {
	if row != speedtest.StartRow || m.speed.Running {
		return nil
	}
	mm, cmd := m.startSpeed()
	*m = mm.(model)
	return cmd
}

// clickPorts maps a pointer position onto the port scan pane: the whole
// action row is the hit area, in either visual form. The pane starts at body
// row 0, aligned with the sidebar.
func (m *model) clickPorts(row int) tea.Cmd {
	if row != portscan.StartRow || m.scan.Phase == portscan.PhaseRunning {
		return nil
	}
	mm, cmd := m.startScan()
	*m = mm.(model)
	return cmd
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
	if x < style.SidebarW+2 || m.sysBusy {
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
