package ui

import (
	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/i18n"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/language"
	"openwrt-monitor-tui/ui/settings/startup"
	"openwrt-monitor-tui/ui/settings/system"
	"openwrt-monitor-tui/ui/settings/wifi"
	"openwrt-monitor-tui/ui/style"

	tea "github.com/charmbracelet/bubbletea"
)

// sections is the settings sidebar. Wi-Fi is absent without a radio.
// resetSettings puts the tab back to its initial UI state: the first section
// highlighted in the sidebar (but not entered, so the keyboard still needs
// Enter), the first band, the first form row, and no pending edits or
// messages. A mouse click inside the pane enters it in the same gesture.
func (m *model) resetSettings() {
	secs := m.sections()
	if len(secs) == 0 {
		return
	}
	m.subOpen = secs[0]
	m.subSel = 0
	m.subActive = false
	m.band = 0
	m.form = wifi.Form{}
	if bands := m.snap.Wifi.Bands; len(bands) > 0 {
		m.form = wifi.FromBand(bands[0], wifi.FocusEnabled)
	}
	m.formNote, m.formErr = "", ""
	m.cron = cron.NewForm()
	m.cronTab = cron.TabList
	m.cronCursor = 0
	m.cronNote, m.cronErr = "", ""
	m.sysSel, m.sysConfirm, m.sysBusy = 0, false, false
	m.sysNote, m.sysErr = "", ""
	m.startup = startup.FromInfo(m.snap.Startup, 0)
	m.startupNote, m.startupErr = "", ""
	m.scroll = 0
	m.langSel = i18n.LangChoice()
	m.langFocus = m.langSel
}

func (m model) sections() []settings.Section {
	return settings.Sections(m.snap.Wifi.Supported)
}

// sectionIndex is the sidebar row of the open section.
func (m model) sectionIndex() int {
	for i, s := range m.sections() {
		if s == m.subOpen {
			return i
		}
	}
	return 0
}

// syncForm reloads the form from the latest sample unless the user has edits in
// flight. It also clamps the band/sub-tab indices when a radio disappears.
func (m *model) syncForm() {
	secs := m.sections()
	if m.subSel >= len(secs) {
		m.subSel = len(secs) - 1
	}
	if m.subSel < 0 {
		m.subSel = 0
	}
	if m.subOpen == settings.Wifi && !m.snap.Wifi.Supported {
		m.subOpen = secs[0]
		m.subSel = m.sectionIndex()
	}
	m.syncWifi()
	m.syncStartup()
	m.syncCron()
}

// syncCron keeps the list cursor inside the job list as the file changes.
func (m *model) syncCron() {
	n := len(m.snap.Cron.Jobs)
	if n == 0 {
		m.cronCursor = 0
		return
	}
	if m.cronCursor >= n {
		m.cronCursor = n - 1
	}
}

func (m *model) syncWifi() {
	bands := m.snap.Wifi.Bands
	if len(bands) == 0 {
		m.form = wifi.Form{}
		return
	}
	if m.band >= len(bands) {
		m.band = len(bands) - 1
	}
	if m.form.Busy || m.form.Editing || m.form.Changed() {
		return
	}
	m.form = wifi.FromBand(bands[m.band], m.form.Focus)
}

// syncStartup follows the file unless the user has unsaved lines in flight.
func (m *model) syncStartup() {
	if m.snap.Startup.Loading {
		return // keep whatever we have until the read lands
	}
	if m.startup.Busy || m.startup.Editing || m.startup.Changed() {
		m.startup.ClampFocus()
		return
	}
	m.startup = startup.FromInfo(m.snap.Startup, m.startup.Focus)
}

// selectBand switches the visible band, but never silently drops edits.
func (m *model) selectBand(i int) {
	bands := m.snap.Wifi.Bands
	if i < 0 || i >= len(bands) || m.form.Busy {
		return
	}
	if m.form.Changed() || m.form.Editing {
		m.formNote = style.Tr.ApplyFirst
		return
	}
	m.band = i
	m.formNote, m.formErr = "", ""
	m.syncForm()
}

// stepBand moves one band left or right in the Wi-Fi pane. It is a no-op when
// the previewed section is not Wi-Fi, or when there is only one band.
func (m *model) stepBand(d int) {
	if m.subOpen != settings.Wifi || len(m.snap.Wifi.Bands) < 2 {
		return
	}
	m.selectBand(wrapIndex(m.band, d, len(m.snap.Wifi.Bands)))
}

func (m model) currentBand() (collect.WifiBand, bool) {
	bands := m.snap.Wifi.Bands
	if len(bands) == 0 {
		return collect.WifiBand{}, false
	}
	return bands[min(m.band, len(bands)-1)], true
}

// selectSubtab moves the sidebar cursor and previews that section. Leaving
// Wi-Fi with unsaved edits is blocked.
func (m *model) selectSubtab(i int) {
	secs := m.sections()
	if i < 0 || i >= len(secs) {
		return
	}
	if secs[i] == m.subOpen {
		m.subSel = i
		return
	}
	if m.subOpen == settings.Wifi &&
		(m.form.Busy || m.form.Editing || m.form.Changed()) {
		m.formNote = style.Tr.ApplyFirst
		return
	}
	m.subSel = i
	m.subOpen = secs[i]
	m.formNote, m.formErr = "", ""
	m.cronNote, m.cronErr = "", ""
	m.sysConfirm = false
	m.scroll = 0
}

func (m *model) moveSubtab(d int) {
	secs := m.sections()
	if len(secs) == 0 {
		return
	}
	m.selectSubtab(wrapIndex(m.subSel, d, len(secs)))
}

// openSubtab enters the previewed section; the cursor already points at it.
func (m *model) openSubtab() {
	secs := m.sections()
	if len(secs) == 0 {
		return
	}
	m.subSel = min(m.subSel, len(secs)-1)
	m.subOpen = secs[m.subSel]
	m.subActive = true
	m.formNote, m.formErr = "", ""
	m.cronNote, m.cronErr = "", ""
	m.sysConfirm = false
	m.scroll = 0
}

func (m model) settingsState() settings.State {
	bands := m.snap.Wifi.Bands
	band := m.band
	if len(bands) > 0 {
		band = min(band, len(bands)-1)
	}
	return settings.State{
		Sections: m.sections(),
		Sel:      m.subSel,
		Open:     m.subOpen,
		Active:   m.subActive,
		Wifi: wifi.State{
			Form:     m.form,
			Bands:    bands,
			Band:     band,
			Active:   m.subActive,
			ShowKeys: m.showKeys,
			Note:     m.formNote,
			Err:      m.formErr,
		},
		Cron: cron.State{
			Form:    m.cron,
			Tab:     m.cronTab,
			Cursor:  m.cronCursor,
			Jobs:    m.snap.Cron.Jobs,
			Avail:   m.snap.Cron.Available,
			Loading: m.snap.Cron.Loading,
			Active:  m.subActive,
			Note:    m.cronNote,
			Err:     m.cronErr,
		},
		System: system.State{
			Sel:     m.sysSel,
			Confirm: m.sysConfirm,
			Busy:    m.sysBusy,
			Active:  m.subActive,
			Note:    m.sysNote,
			Err:     m.sysErr,
		},
		Startup: startup.State{
			Form:    m.startup,
			Path:    startupPath(m.snap.Startup.Path),
			Active:  m.subActive,
			Note:    m.startupNote,
			Err:     m.startupErr,
			Loading: m.snap.Startup.Loading,
		},
		LangSel:   m.langSel,
		LangFocus: m.langFocus,
	}
}

// activate runs whatever Enter/Space means for the focused row.
func (m model) activate() (tea.Model, tea.Cmd) {
	f := &m.form
	if f.Busy {
		return m, nil
	}
	m.formNote, m.formErr = "", ""

	switch f.Focus {
	case wifi.FocusEnabled:
		f.Enabled.Toggle()
	case wifi.FocusSSID, wifi.FocusPassword:
		if f.Iface == "" {
			m.formNote = style.Tr.NoAP
			return m, nil
		}
		f.Text().Begin()
		f.Editing = true
	case wifi.FocusChannel, wifi.FocusSecurity:
		// Arrows only change a value once the row has been entered, so
		// Enter is what turns the row into an editor.
		if f.Focus == wifi.FocusSecurity && f.Iface == "" {
			m.formNote = style.Tr.NoAP
			return m, nil
		}
		f.Picker().Begin()
		f.Editing = true
	case wifi.FocusApply:
		return m.applyForm()
	case wifi.FocusCancel:
		f.Reset()
		m.formNote = style.Tr.Reverted
	}
	return m, nil
}

// toggleButtonFocus moves between the two buttons with left/right.
func (m *model) toggleButtonFocus() bool {
	switch m.form.Focus {
	case wifi.FocusApply:
		m.form.Focus = wifi.FocusCancel
		return true
	case wifi.FocusCancel:
		m.form.Focus = wifi.FocusApply
		return true
	}
	return false
}

func (m model) applyForm() (tea.Model, tea.Cmd) {
	f := &m.form
	if f.Busy {
		return m, nil
	}
	if msg := f.Validate(); msg != "" {
		m.formErr = msg
		return m, nil
	}
	sets := f.Plan()
	if len(sets) == 0 {
		m.formNote = style.Tr.NoChanges
		return m, nil
	}
	f.Busy = true
	return m, m.applyCmd(sets)
}

func (m model) applyCmd(sets []wifi.Set) tea.Cmd {
	return func() tea.Msg { return wifiApplyMsg{err: wifi.Apply(sets)} }
}

// updateEdit swallows every key while a text field is being edited, so "q"
// types a q instead of quitting and Esc only leaves the field.
func (m model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.subOpen {
	case settings.Cron:
		return m.updateCronEdit(msg)
	case settings.Startup:
		return m.updateStartupEdit(msg)
	}
	switch msg.Type {
	case tea.KeyEsc:
		if in := m.form.Text(); in != nil {
			in.Cancel()
		}
		if sel := m.form.Picker(); sel != nil {
			sel.Cancel()
		}
		m.form.Editing = false
		return m, nil
	case tea.KeyEnter:
		m.form.CommitEdit()
		return m, nil
	}

	// Only the widget owning the row sees the remaining keys.
	if in := m.form.Text(); in != nil {
		switch msg.Type {
		case tea.KeyCtrlU:
			in.Clear()
		case tea.KeyBackspace:
			in.Backspace()
		case tea.KeySpace:
			in.Type(' ')
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				if r >= ' ' && r != 0x7f {
					in.Type(r)
				}
			}
			m.formErr, m.formNote = "", ""
		}
		return m, nil
	}
	if sel := m.form.Picker(); sel != nil {
		switch msg.Type {
		case tea.KeyLeft:
			sel.Step(-1)
		case tea.KeyRight:
			sel.Step(+1)
		}
		m.formErr, m.formNote = "", ""
	}
	return m, nil
}

func (m *model) chooseLang(i int) {
	if i < 0 || i >= len(language.Choices) {
		return
	}
	m.langFocus, m.langSel = i, i
	i18n.SetLangChoice(language.Choices[i])
	// The cron pickers show localized labels; rebuild them for the new
	// language. There is no unsaved job in flight when the picker is used.
	m.cron = cron.NewForm()
	m.cronNote, m.cronErr = "", ""
}

// editing reports whether a settings form currently owns the keyboard.
func (m model) editing() bool {
	if m.tab != tabSettings {
		return false
	}
	switch m.subOpen {
	case settings.Cron:
		return m.cron.Editing
	case settings.Startup:
		return m.startup.Editing
	}
	return m.form.Editing
}

// commitEdit closes the active row, keeping its value.
func (m *model) commitEdit() {
	switch m.subOpen {
	case settings.Cron:
		m.cron.CommitEdit()
	case settings.Startup:
		m.startup.CommitEdit()
	default:
		m.form.CommitEdit()
	}
}

// updateCron handles the job list and the edit form behind it.
func (m model) updateCron(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.cronTab == cron.TabList {
		return m.updateCronList(msg)
	}
	f := &m.cron
	switch msg.String() {
	case "up", "k":
		f.StepFocus(-1)
	case "down", "j":
		f.StepFocus(+1)
	case "left", "h", "right", "l":
		// The buttons own the arrows; elsewhere they switch sub-tabs.
		if !m.cronButtons() {
			m.cronTab = cron.TabList
		}
	case "enter", " ":
		return m.activateCron()
	case "n":
		// Start over with an empty job without leaving the editor.
		m.cron = cron.NewForm()
		m.cronNote, m.cronErr = "", ""
	}
	return m, nil
}

func (m model) updateCronList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	jobs := m.snap.Cron.Jobs
	switch msg.String() {
	case "up", "k":
		m.cronCursor = wrapIndex(m.cronCursor, -1, len(jobs))
	case "down", "j":
		m.cronCursor = wrapIndex(m.cronCursor, +1, len(jobs))
	case "enter":
		m.cronLoad()
	case " ":
		if job, ok := m.cronSelected(); ok {
			return m, m.cronToggle(job)
		}
	case "n":
		m.cron = cron.NewForm()
		m.cronTab = cron.TabEditor
		m.cronNote, m.cronErr = "", ""
	case "left", "h", "right", "l":
		m.cronTab = cron.TabEditor
	}
	return m, nil
}

// cronSelected is the job under the list cursor.
func (m model) cronSelected() (collect.CronJob, bool) {
	jobs := m.snap.Cron.Jobs
	if len(jobs) == 0 {
		return collect.CronJob{}, false
	}
	if m.cronCursor >= len(jobs) {
		m.cronCursor = len(jobs) - 1
	}
	return jobs[m.cronCursor], true
}

// cronToggle flips the switch and writes the crontab in the background. The
// list is updated in place so the switch reacts immediately.
func (m *model) cronToggle(job collect.CronJob) tea.Cmd {
	on := !job.Enabled
	for i := range m.snap.Cron.Jobs {
		if m.snap.Cron.Jobs[i].Line == job.Line {
			m.snap.Cron.Jobs[i].Enabled = on
		}
	}
	return func() tea.Msg { return cronToggleMsg{line: job.Line, err: cron.SetEnabled(job, on)} }
}

// cronLoad opens the selected job in the form for editing.
func (m *model) cronLoad() {
	jobs := m.snap.Cron.Jobs
	if len(jobs) == 0 {
		return
	}
	if m.cronCursor >= len(jobs) {
		m.cronCursor = len(jobs) - 1
	}
	m.cron.Load(jobs[m.cronCursor])
	m.cronTab = cron.TabEditor
	m.cronNote, m.cronErr = "", ""
}

// cronButtons flips between Apply and Cancel. The other rows only change once
// Enter has entered them, so arrows do nothing there.
func (m *model) cronButtons() bool {
	switch m.cron.Focus {
	case cron.FocusApply, cron.FocusCancel:
		m.cron.Focus = cron.FocusApply + cron.FocusCancel - m.cron.Focus
		return true
	}
	return false
}

func (m model) activateCron() (tea.Model, tea.Cmd) {
	f := &m.cron
	if f.Busy {
		return m, nil
	}
	m.cronNote, m.cronErr = "", ""

	switch f.Focus {
	case cron.FocusKind, cron.FocusWeekday, cron.FocusWhen, cron.FocusCommand:
		if sel := f.Picker(); sel != nil {
			sel.Begin()
		}
		if in := f.Text(); in != nil {
			in.Begin()
		}
		f.Editing = true
	case cron.FocusApply:
		if msg := f.Validate(); msg != "" {
			m.cronErr = msg
			return m, nil
		}
		if f.EditLine >= 0 && !f.Changed() {
			m.cronNote = style.Tr.NoChanges
			return m, nil
		}
		sched, _ := f.Schedule()
		line := f.EditLine
		origin := f.Origin()
		command := f.CommandLine()
		f.Busy = true
		if line >= 0 {
			return m, func() tea.Msg {
				return cronApplyMsg{err: cron.Update(line, origin, sched, command), updated: true}
			}
		}
		return m, func() tea.Msg {
			return cronApplyMsg{err: cron.Add(sched, command)}
		}
	case cron.FocusCancel:
		// Cancel ends the whole edit session, so the tab goes back to "new".
		m.cron = cron.NewForm()
		m.cronTab = cron.TabList
		m.cronNote = style.Tr.Reverted
	}
	return m, nil
}

// updateCronEdit keeps every key inside the entered row.
func (m model) updateCronEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		if in := m.cron.Text(); in != nil {
			in.Cancel()
		}
		if sel := m.cron.Picker(); sel != nil {
			sel.Cancel()
		}
		m.cron.Editing = false
		return m, nil
	case tea.KeyEnter:
		m.cron.CommitEdit()
		return m, nil
	}

	if in := m.cron.Text(); in != nil {
		switch msg.Type {
		case tea.KeyCtrlU:
			in.Clear()
		case tea.KeyBackspace:
			in.Backspace()
		case tea.KeySpace:
			in.Type(' ')
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				if r >= ' ' && r != 0x7f {
					in.Type(r)
				}
			}
			m.cronErr, m.cronNote = "", ""
		}
		return m, nil
	}
	if sel := m.cron.Picker(); sel != nil {
		switch msg.Type {
		case tea.KeyLeft:
			sel.Step(-1)
		case tea.KeyRight:
			sel.Step(+1)
		}
		m.cronErr, m.cronNote = "", ""
	}
	return m, nil
}

// updateStartup handles the rc.local line editor.
func (m model) updateStartup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := &m.startup
	switch msg.String() {
	case "up", "k":
		f.Focus = wrapIndex(f.Focus, -1, f.FocusCount())
		m.ensureStartupVisible()
	case "down", "j":
		f.Focus = wrapIndex(f.Focus, +1, f.FocusCount())
		m.ensureStartupVisible()
	case " ":
		if f.Available && f.Focus < f.TaskCount() {
			f.Toggle(f.Focus)
		}
	case "a":
		if f.Available {
			f.Insert()
			m.ensureStartupVisible()
		}
	case "d":
		if f.Available && f.Focus < f.TaskCount() {
			f.Delete(f.Focus)
			m.ensureStartupVisible()
		}
	case "enter":
		return m.activateStartup()
	}
	return m, nil
}

// ensureStartupVisible scrolls the body so the focused line stays on screen.
func (m *model) ensureStartupVisible() {
	row := startup.RowOffset(m.settingsState().Startup)
	if row < m.scroll {
		m.scroll = row
	}
	if row >= m.scroll+m.rows() {
		m.scroll = row - m.rows() + 1
	}
}

func (m model) activateStartup() (tea.Model, tea.Cmd) {
	f := &m.startup
	if f.Busy || !f.Available {
		return m, nil
	}
	m.startupNote, m.startupErr = "", ""

	switch f.Focus {
	case f.ApplyIndex():
		if !f.Changed() {
			m.startupNote = style.Tr.NoChanges
			return m, nil
		}
		f.Busy = true
		body := f.Body()
		return m, func() tea.Msg { return startupApplyMsg{err: startup.Save(body)} }
	case f.CancelIndex():
		f.Reset()
		m.startupNote = style.Tr.Reverted
	default:
		f.Begin(f.Focus)
	}
	return m, nil
}

// updateStartupEdit keeps every key inside the line being edited.
func (m model) updateStartupEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := &m.startup
	switch msg.Type {
	case tea.KeyEsc:
		f.Editing = false
		return m, nil
	case tea.KeyEnter:
		f.CommitEdit()
		return m, nil
	}

	in := f.Edit()
	switch msg.Type {
	case tea.KeyCtrlU:
		in.Clear()
	case tea.KeyBackspace:
		in.Backspace()
	case tea.KeySpace:
		in.Type(' ')
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= ' ' && r != 0x7f {
				in.Type(r)
			}
		}
		m.startupErr, m.startupNote = "", ""
	}
	return m, nil
}

// startupPath falls back to the well-known location when a sample has none.
func startupPath(p string) string {
	if p != "" {
		return p
	}
	return collect.StartupPath()
}

// updateSystem drives the reboot/power-off/factory-reset page. Enter asks for
// confirmation first; a second Enter runs the action.
func (m model) updateSystem(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.sysBusy {
		return m, nil
	}
	switch msg.String() {
	case "up", "k":
		m.sysSel = wrapIndex(m.sysSel, -1, int(system.Count))
		m.sysConfirm = false
	case "down", "j":
		m.sysSel = wrapIndex(m.sysSel, +1, int(system.Count))
		m.sysConfirm = false
	case "enter", " ":
		if !m.sysConfirm {
			m.sysConfirm = true
			m.sysNote, m.sysErr = "", ""
			return m, nil
		}
		action := system.Action(m.sysSel)
		m.sysBusy = true
		return m, func() tea.Msg { return systemApplyMsg{err: system.Run(action)} }
	}
	return m, nil
}
