package ui

import (
	"fmt"
	"openwrt-monitor-tui/ui/i18n"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/startup"
	"openwrt-monitor-tui/ui/settings/system"
	"openwrt-monitor-tui/ui/style"
)

// openSection previews and enters a settings section the way the keyboard does.
func openSection(m model, sec settings.Section) model {
	for i, s := range m.sections() {
		if s == sec {
			m.subSel = i
			m.selectSubtab(i)
			m.openSubtab()
			return m
		}
	}
	return m
}

func TestCronSectionExists(t *testing.T) {
	i18n.SetLang("en")
	m := newTestModel()
	secs := m.sections()
	want := []settings.Section{settings.Wifi, settings.Cron, settings.Startup, settings.System, settings.Language}
	if len(secs) != len(want) {
		t.Fatalf("sections: %v", secs)
	}
	for i := range want {
		if secs[i] != want[i] {
			t.Fatalf("section %d: %v", i, secs[i])
		}
	}
	if settings.Name(settings.Cron) != style.Tr.Cron || settings.Name(settings.System) != style.Tr.System {
		t.Fatalf("names: %q %q", settings.Name(settings.Cron), settings.Name(settings.System))
	}
}

// cronModel opens the cron section on the list tab with the given jobs.
func cronModel(jobs ...collect.CronJob) model {
	m := newTestModel()
	m.snap.Cron = collect.CronInfo{Available: true, Jobs: jobs}
	m.cronTab = cron.TabList
	m.syncForm()
	return openSection(m, settings.Cron)
}

func jobList() []collect.CronJob {
	return []collect.CronJob{
		{Schedule: "*/5 * * * *", Command: "/usr/bin/logger hi", Line: 1, Enabled: true},
		{Schedule: "0 3 * * *", Command: "/etc/init.d/dropbear restart", Line: 3, Enabled: true},
	}
}

func TestCronListShowsCommandsOnly(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	m.subActive = true
	body := strings.Join(m.body(), "\n")
	if !strings.Contains(body, "/usr/bin/logger hi") || !strings.Contains(body, "/etc/init.d/dropbear restart") {
		t.Fatalf("commands missing:\n%s", body)
	}
	if strings.Contains(body, "*/5") || strings.Contains(body, "0 3 * *") {
		t.Fatalf("schedule leaked into the list:\n%s", body)
	}
	if !strings.Contains(body, "[JOBS]") && !strings.Contains(body, "JOBS") {
		t.Fatalf("tab strip missing:\n%s", body)
	}
}

func TestCronEditExistingJob(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	key := func(s string) tea.KeyMsg {
		switch s {
		case "enter":
			return tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			return tea.KeyMsg{Type: tea.KeyEsc}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}

	// Enter on the list loads the job into the form
	mm, _ := m.Update(key("enter"))
	m = mm.(model)
	if m.cronTab != cron.TabEditor || m.cron.EditLine != 1 {
		t.Fatalf("load: tab=%d line=%d", m.cronTab, m.cron.EditLine)
	}
	if m.cron.Command.Value != "/usr/bin/logger hi" {
		t.Fatalf("command: %q", m.cron.Command.Value)
	}
	if m.cron.KindIndex() != cron.KindMinutes || m.cron.When.Value != "5" {
		t.Fatalf("schedule: kind=%d when=%q", m.cron.KindIndex(), m.cron.When.Value)
	}
	if m.cron.Changed() {
		t.Fatal("freshly loaded job reported as changed")
	}

	// change only the command
	m.cron.Focus = cron.FocusCommand
	mm, _ = m.Update(key("enter"))
	m = mm.(model)
	m.cron.Text().Clear()
	for _, r := range "/usr/bin/logger bye" {
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mm.(model)
	}
	mm, _ = m.Update(key("enter"))
	m = mm.(model)

	m.cron.Focus = cron.FocusApply
	mm, cmd := m.Update(key("enter"))
	m = mm.(model)
	if cmd == nil || !m.cron.Busy {
		t.Fatalf("apply not dispatched: cmd=%v busy=%v err=%q", cmd != nil, m.cron.Busy, m.cronErr)
	}

	mm, _ = m.Update(cronApplyMsg{updated: true})
	m = mm.(model)
	if m.cronTab != cron.TabList || m.cronNote != style.Tr.CronUpdated || m.cron.EditLine != -1 {
		t.Fatalf("after update: tab=%d note=%q line=%d", m.cronTab, m.cronNote, m.cron.EditLine)
	}
}

func TestCronUpdateChangesSchedule(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)

	// switch the loaded job to "daily at 08:30"
	m.cron.Focus = cron.FocusKind
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if !m.cron.Editing {
		t.Fatal("kind row did not open")
	}
	for i := 0; i < 2; i++ { // minutes -> hours -> daily
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = mm.(model)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)

	m.cron.Focus = cron.FocusWhen
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	m.cron.Text().Clear()
	for _, r := range "08:30" {
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mm.(model)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)

	if sched, msg := m.cron.Schedule(); sched != "30 8 * * *" || msg != "" {
		t.Fatalf("schedule: %q %q", sched, msg)
	}
	if !m.cron.Changed() {
		t.Fatal("schedule change not detected")
	}
}

func TestCronCancelEndsSession(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if m.cron.EditLine < 0 {
		t.Fatal("job not loaded")
	}
	m.cron.Command.Value = "changed"
	m.cron.Focus = cron.FocusCancel
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if m.cronTab != cron.TabList || m.cron.EditLine != -1 || m.cron.Command.Value != "" {
		t.Fatalf("cancel: tab=%d line=%d cmd=%q", m.cronTab, m.cron.EditLine, m.cron.Command.Value)
	}
	if m.cronNote != style.Tr.Reverted {
		t.Fatalf("note: %q", m.cronNote)
	}
}

func TestCronNewJob(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)

	// "n" starts a fresh job in the editor
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	m = mm.(model)
	if m.cronTab != cron.TabEditor || m.cron.EditLine != -1 {
		t.Fatalf("new: tab=%d line=%d", m.cronTab, m.cron.EditLine)
	}

	// an empty job fails validation before any write
	m.cron.Focus = cron.FocusApply
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if cmd != nil || m.cron.Busy || m.cronErr != style.Tr.ValCronRange {
		t.Fatalf("invalid apply: cmd=%v busy=%v err=%q", cmd != nil, m.cron.Busy, m.cronErr)
	}

	m.cron.When.Value = "15"
	m.cron.Command.Value = "/bin/true"
	m.cron.Focus = cron.FocusApply
	mm, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if cmd == nil || !m.cron.Busy {
		t.Fatal("new job not dispatched")
	}
	mm, _ = m.Update(cronApplyMsg{})
	m = mm.(model)
	if m.cronTab != cron.TabList || m.cronNote != style.Tr.CronAdded {
		t.Fatalf("after add: tab=%d note=%q", m.cronTab, m.cronNote)
	}
}

func TestCronClicks(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	m.subActive = true

	// sub-tab strip: second tab opens the editor
	hits := cron.TabHits(m.settingsState().Cron)
	x := (hits[1][0] + hits[1][1]) / 2
	mm, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x, Y: 4})
	m = mm.(model)
	if m.cronTab != cron.TabEditor {
		t.Fatalf("tab click: %d", m.cronTab)
	}

	// back to the list, then click the second job to edit it
	x0 := (hits[0][0] + hits[0][1]) / 2
	mm, _ = m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x0, Y: 4})
	m = mm.(model)
	if m.cronTab != cron.TabList {
		t.Fatalf("tab click back: %d", m.cronTab)
	}
	st := m.settingsState().Cron
	y := 4 + cron.ContentOffset(st) + 1
	mm, _ = m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: style.SidebarW + 14, Y: y})
	m = mm.(model)
	if m.cronTab != cron.TabEditor || m.cron.EditLine != 3 {
		t.Fatalf("row click: tab=%d line=%d", m.cronTab, m.cron.EditLine)
	}

	// clicking a text row starts editing it
	off := cron.FormRowOffset(m.settingsState().Cron)
	mm, _ = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: style.SidebarW + 6, Y: 4 + off + cron.FocusCommand,
	})
	got := mm.(model)
	if !got.cron.Editing || got.cron.Focus != cron.FocusCommand {
		t.Fatalf("form click: editing=%v focus=%d", got.cron.Editing, got.cron.Focus)
	}
}

type errString string

func (e errString) Error() string { return string(e) }

// startupModel opens the startup editor with n task lines plus the usual
// shebang/comment/exit furniture.
func startupModel(lines int) model {
	m := newTestModel()
	content := []string{"#!/bin/sh", "# comment"}
	for i := 0; i < lines; i++ {
		content = append(content, fmt.Sprintf("/bin/task%d", i))
	}
	content = append(content, "exit 0")
	m.snap.Startup = collect.StartupInfo{Available: true, Path: "/etc/rc.local", Content: content}
	m.syncForm()
	return openSection(m, settings.Startup)
}

func TestStartupNavigation(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(3)
	if m.subOpen != settings.Startup || !m.subActive {
		t.Fatalf("not in startup: open=%v active=%v", m.subOpen, m.subActive)
	}
	if m.startup.TaskCount() != 3 || m.startup.Focus != 0 {
		t.Fatalf("form: tasks=%d focus=%d", m.startup.TaskCount(), m.startup.Focus)
	}

	// j walks the tasks, then the buttons, then wraps
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = mm.(model)
	if m.startup.Focus != 1 {
		t.Fatalf("focus: %d", m.startup.Focus)
	}
	for i := 0; i < 2; i++ {
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		m = mm.(model)
	}
	if m.startup.Focus != m.startup.ApplyIndex() {
		t.Fatalf("apply focus: %d", m.startup.Focus)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m = mm.(model)
	if m.startup.Focus != 4 {
		t.Fatalf("cancel focus: %d", m.startup.Focus)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if mm.(model).startup.Focus != 0 {
		t.Fatalf("wrap: %d", mm.(model).startup.Focus)
	}

	// a long file keeps the cursor on screen
	long := startupModel(40)
	long.startup.Focus = 30
	long.ensureStartupVisible()
	row := startup.RowOffset(long.settingsState().Startup)
	if row < long.scroll || row >= long.scroll+long.rows() {
		t.Fatalf("cursor off screen: row=%d scroll=%d rows=%d", row, long.scroll, long.rows())
	}
}

func TestStartupEditing(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(2)

	// enter starts editing the focused task
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if !m.startup.Editing {
		t.Fatal("enter did not open the task")
	}
	m.startup.Edit().Clear()
	for _, r := range "/bin/edited" {
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mm.(model)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	first, _ := m.startup.TaskAt(0)
	if m.startup.Editing || first.Text != "/bin/edited" || !m.startup.Changed() {
		t.Fatalf("commit: editing=%v first=%+v", m.startup.Editing, first)
	}

	// esc drops the edit
	m.startup.Focus = 1
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")})
	m = mm.(model)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = mm.(model)
	second, _ := m.startup.TaskAt(1)
	if m.startup.Editing || second.Text != "/bin/task1" {
		t.Fatalf("esc: %+v", second)
	}

	// a inserts and edits a new task
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	m = mm.(model)
	if m.startup.TaskCount() != 3 || m.startup.Focus != 2 || !m.startup.Editing {
		t.Fatalf("add: tasks=%d focus=%d", m.startup.TaskCount(), m.startup.Focus)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)

	// d removes it again
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	m = mm.(model)
	if m.startup.TaskCount() != 2 {
		t.Fatalf("delete: %d", m.startup.TaskCount())
	}
}

func TestStartupToggle(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(2)
	m.subActive = true

	// space flips the switch and marks the form dirty
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mm.(model)
	first, _ := m.startup.TaskAt(0)
	if first.Enabled || !m.startup.Changed() {
		t.Fatalf("toggle: %+v changed=%v", first, m.startup.Changed())
	}
	if !strings.Contains(m.startup.Body(), "#OFF# /bin/task0") {
		t.Fatalf("body:\n%s", m.startup.Body())
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mm.(model)
	if first, _ := m.startup.TaskAt(0); !first.Enabled || m.startup.Changed() {
		t.Fatalf("toggle back: %+v", first)
	}
}

func TestStartupApplyAndCancel(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(2)
	m.startup.Toggle(1)

	m.startup.Focus = m.startup.ApplyIndex()
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if cmd == nil || !m.startup.Busy {
		t.Fatalf("apply not dispatched: cmd=%v busy=%v err=%q", cmd != nil, m.startup.Busy, m.startupErr)
	}

	mm, _ = m.Update(startupApplyMsg{})
	m = mm.(model)
	if m.startup.Busy || m.startupErr != "" || m.startupNote != style.Tr.StartupSaved || m.startup.Changed() {
		t.Fatalf("after save: busy=%v err=%q note=%q changed=%v",
			m.startup.Busy, m.startupErr, m.startupNote, m.startup.Changed())
	}

	// a failure keeps the edits and shows the error
	m.startup.Toggle(0)
	m.startup.Busy = true
	mm, _ = m.Update(startupApplyMsg{err: errString("read-only file system")})
	m = mm.(model)
	if m.startup.Busy || !strings.Contains(m.startupErr, "read-only") || !m.startup.Changed() {
		t.Fatalf("failure: busy=%v err=%q", m.startup.Busy, m.startupErr)
	}

	// Cancel restores the file
	m.startup.Focus = m.startup.CancelIndex()
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if m.startup.Changed() {
		t.Fatal("cancel did not restore")
	}
}

func TestStartupEditsSurviveSamples(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(2)
	m.startup.Toggle(0)
	m.subActive = true
	m.syncForm() // as if a sample arrived
	first, _ := m.startup.TaskAt(0)
	if first.Enabled {
		t.Fatalf("toggle lost: %+v", first)
	}

	// a clean form does follow the file
	m.startup.Reset()
	m.snap.Startup.Content = []string{"#!/bin/sh", "/bin/new", "exit 0"}
	m.syncForm()
	first, _ = m.startup.TaskAt(0)
	if first.Text != "/bin/new" {
		t.Fatalf("refresh: %+v", first)
	}
}

func TestStartupClicks(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(2)
	m.subActive = true
	off := startup.LineOffset(m.settingsState().Startup)

	// clicking the switch flips it; clicking the text edits it.
	// The switch is drawn at columns 4..6 of the pane content.
	check := [2]int{4, 6}
	mm, _ := m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: style.SidebarW + 2 + check[0], Y: 4 + off,
	})
	m = mm.(model)
	if first, _ := m.startup.TaskAt(0); first.Enabled || m.startup.Focus != 0 {
		t.Fatalf("switch click: %+v focus=%d", first, m.startup.Focus)
	}

	mm, _ = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: style.SidebarW + 12, Y: 4 + off + 1,
	})
	m = mm.(model)
	if !m.startup.Editing || m.startup.Focus != 1 {
		t.Fatalf("text click: editing=%v focus=%d", m.startup.Editing, m.startup.Focus)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = mm.(model)

	// Apply button
	off = startup.LineOffset(m.settingsState().Startup)
	row := off + max(1, m.startup.TaskCount()) + 1
	btns := m.startup.Buttons()
	ap := btns.Hits()[0]
	mm, cmd := m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: style.SidebarW + 2 + ap[0], Y: 4 + row,
	})
	got := mm.(model)
	if cmd == nil || !got.startup.Busy {
		t.Fatalf("apply click: cmd=%v busy=%v", cmd != nil, got.startup.Busy)
	}
}

func TestCronToggle(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	m.subActive = true

	// the list shows a switch per job
	body := strings.Join(m.body(), "\n")
	if !strings.Contains(body, "[x] /usr/bin/logger hi") {
		t.Fatalf("switch missing:\n%s", body)
	}

	// space flips the selected job
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mm.(model)
	if cmd == nil {
		t.Fatal("toggle not dispatched")
	}
	mm, _ = m.Update(cronToggleMsg{})
	m = mm.(model)
	if m.cronErr != "" {
		t.Fatalf("toggle error: %q", m.cronErr)
	}

	// clicking the switch also toggles; the switch is drawn at columns 4..6
	check := [2]int{4, 6}
	st := m.settingsState().Cron
	mm, cmd = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: style.SidebarW + 2 + check[0], Y: 4 + cron.ContentOffset(st),
	})
	if cmd == nil {
		t.Fatal("switch click did not toggle")
	}
}

func TestCronLoadingState(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	m.snap.Cron.Loading = true
	m.subActive = true

	body := strings.Join(m.body(), "\n")
	if !strings.Contains(body, style.Tr.Loading) {
		t.Fatalf("loading placeholder missing:\n%s", body)
	}
	if strings.Contains(body, "/usr/bin/logger hi") {
		t.Fatalf("stale jobs shown while loading:\n%s", body)
	}

	// the editor tab does not depend on the file
	m.cronTab = cron.TabEditor
	body = strings.Join(m.body(), "\n")
	if strings.Contains(body, style.Tr.Loading) || !strings.Contains(body, style.Tr.CronKind) {
		t.Fatalf("editor while loading:\n%s", body)
	}
}

func TestStartupLoadingState(t *testing.T) {
	i18n.SetLang("en")
	m := startupModel(2)
	m.snap.Startup.Loading = true
	m.syncForm() // must keep the form that is already loaded
	m.subActive = true

	body := strings.Join(m.body(), "\n")
	if !strings.Contains(body, style.Tr.Loading) || strings.Contains(body, "/bin/task0") {
		t.Fatalf("loading pane:\n%s", body)
	}
	if m.startup.TaskCount() != 2 {
		t.Fatalf("form was clobbered: %d tasks", m.startup.TaskCount())
	}

	m.snap.Startup.Loading = false
	m.syncForm()
	body = strings.Join(m.body(), "\n")
	if strings.Contains(body, style.Tr.Loading) || !strings.Contains(body, "/bin/task0") {
		t.Fatalf("after load:\n%s", body)
	}
}

func TestNewSkipsJobFiles(t *testing.T) {
	i18n.SetLang("en")
	m := New(time.Second, time.Now()).(model)
	if m.snap.Cron.Path != "" || m.snap.Startup.Path != "" {
		t.Fatalf("startup parsed the job files: cron=%q startup=%q", m.snap.Cron.Path, m.snap.Startup.Path)
	}
	if m.snap.Cron.Loading || m.snap.Startup.Loading {
		t.Fatal("boot sample started the background reader")
	}
}

func TestJobsReloadMessage(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	mm, _ := m.Update(jobsReloadMsg{
		cron: collect.CronInfo{Available: true, Path: "/etc/crontabs/root", Jobs: []collect.CronJob{
			{Schedule: "*/9 * * * *", Command: "/bin/new", Line: 0, Enabled: true},
		}},
		startup: collect.StartupInfo{Available: true, Path: "/etc/rc.local", Content: []string{"#!/bin/sh", "/bin/boot", "exit 0"}},
	})
	m = mm.(model)
	if len(m.snap.Cron.Jobs) != 1 || m.snap.Cron.Jobs[0].Command != "/bin/new" {
		t.Fatalf("cron not applied: %+v", m.snap.Cron)
	}
	if m.startup.TaskCount() != 1 {
		t.Fatalf("startup not applied: %d tasks", m.startup.TaskCount())
	}
}

func TestCronToggleRequestsReload(t *testing.T) {
	i18n.SetLang("en")
	m := cronModel(jobList()...)
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = mm.(model)
	if cmd == nil {
		t.Fatal("toggle did not return a command")
	}
	// the switch flips immediately, before the write finishes
	if job := m.snap.Cron.Jobs[0]; job.Enabled {
		t.Fatalf("job still enabled: %+v", job)
	}
}

func TestSettingsArrowScope(t *testing.T) {
	i18n.SetLang("en")
	base := newTestModel()
	base.tab = 0
	base.setTab(tabSettings)
	base.band = 1 // would be touched if arrows still belonged to the bands

	// sidebar: arrows keep switching the main tabs
	mm, _ := base.Update(tea.KeyMsg{Type: tea.KeyRight})
	got := mm.(model)
	if got.tab != 0 || got.band != 1 {
		t.Fatalf("sidebar right: tab=%d band=%d", got.tab, got.band)
	}
	mm, _ = base.Update(tea.KeyMsg{Type: tea.KeyLeft})
	got = mm.(model)
	if got.tab != 2 || got.band != 1 {
		t.Fatalf("sidebar left: tab=%d band=%d", got.tab, got.band)
	}

	// language and startup have no horizontal control, so arrows still switch tabs
	for _, sec := range []settings.Section{settings.Language, settings.Startup} {
		m := openSection(newTestModel(), sec)
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		got = mm.(model)
		if got.tab != 0 {
			t.Fatalf("%v: right should switch tabs, tab=%d", sec, got.tab)
		}
	}

	// the Wi-Fi pane owns them once entered
	m := openSection(newTestModel(), settings.Wifi)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	got = mm.(model)
	if got.tab != tabSettings || got.band != 1 {
		t.Fatalf("wifi right: tab=%d band=%d", got.tab, got.band)
	}

	// the cron editor keeps its own sub-tabs
	c := openSection(newTestModel(), settings.Cron)
	c.cronTab = cron.TabEditor
	mm, _ = c.Update(tea.KeyMsg{Type: tea.KeyRight})
	got = mm.(model)
	if got.tab != tabSettings || got.cronTab != cron.TabList {
		t.Fatalf("cron right: tab=%d cronTab=%d", got.tab, got.cronTab)
	}
}

func TestSystemSection(t *testing.T) {
	i18n.SetLang("en")
	m := openSection(newTestModel(), settings.System)
	if !m.subActive || m.subOpen != settings.System {
		t.Fatalf("not in system: open=%v active=%v", m.subOpen, m.subActive)
	}
	body := strings.Join(m.body(), "\n")
	for _, want := range []string{style.Tr.SysReboot, style.Tr.SysPoweroff, style.Tr.SysFactory, style.Tr.SysHint} {
		if !strings.Contains(body, want) {
			t.Fatalf("%q missing:\n%s", want, body)
		}
	}

	key := func(s string) tea.KeyMsg {
		switch s {
		case "enter":
			return tea.KeyMsg{Type: tea.KeyEnter}
		case "esc":
			return tea.KeyMsg{Type: tea.KeyEsc}
		}
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}

	// j selects power off
	mm, _ := m.Update(key("j"))
	m = mm.(model)
	if m.sysSel != 1 || m.sysConfirm {
		t.Fatalf("select: sel=%d confirm=%v", m.sysSel, m.sysConfirm)
	}

	// first Enter asks, and Esc backs out of the question only
	mm, cmd := m.Update(key("enter"))
	m = mm.(model)
	if cmd != nil || !m.sysConfirm {
		t.Fatalf("confirm: cmd=%v confirm=%v", cmd != nil, m.sysConfirm)
	}
	mm, _ = m.Update(key("esc"))
	m = mm.(model)
	if m.sysConfirm || !m.subActive {
		t.Fatalf("esc should only cancel: confirm=%v active=%v", m.sysConfirm, m.subActive)
	}

	// two Enters dispatch the action
	mm, _ = m.Update(key("enter"))
	m = mm.(model)
	mm, cmd = m.Update(key("enter"))
	m = mm.(model)
	if cmd == nil || !m.sysBusy {
		t.Fatalf("run not dispatched: cmd=%v busy=%v", cmd != nil, m.sysBusy)
	}

	// a refusal is reported and the page stays usable
	mm, _ = m.Update(systemApplyMsg{err: errString("operation not permitted")})
	m = mm.(model)
	if m.sysBusy || !strings.Contains(m.sysErr, "not permitted") || m.sysConfirm {
		t.Fatalf("after failure: busy=%v err=%q", m.sysBusy, m.sysErr)
	}

	// leaving the section and coming back resets the selection
	m.setTab(0)
	m.setTab(tabSettings)
	m = openSection(m, settings.System)
	if m.sysSel != 0 || m.sysConfirm || m.sysErr != "" {
		t.Fatalf("reset: sel=%d confirm=%v err=%q", m.sysSel, m.sysConfirm, m.sysErr)
	}
}

func TestSystemClicks(t *testing.T) {
	i18n.SetLang("en")
	m := openSection(newTestModel(), settings.System)
	m.subActive = true
	x := style.SidebarW + 6

	// click the factory row: selects it and asks for confirmation
	mm, cmd := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: x, Y: 4 + system.LineOffset() + 2})
	m = mm.(model)
	if cmd != nil || m.sysSel != 2 || !m.sysConfirm {
		t.Fatalf("click row: sel=%d confirm=%v cmd=%v", m.sysSel, m.sysConfirm, cmd != nil)
	}

	// clicking it again runs it
	mm, cmd = m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: x, Y: 4 + system.LineOffset() + 2})
	m = mm.(model)
	if cmd == nil || !m.sysBusy {
		t.Fatalf("second click: cmd=%v busy=%v", cmd != nil, m.sysBusy)
	}
}
