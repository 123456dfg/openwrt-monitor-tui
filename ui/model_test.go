package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/network"
	"openwrt-monitor-tui/ui/network/speedtest"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/wifi"
	"openwrt-monitor-tui/ui/status"
	"openwrt-monitor-tui/ui/widgets/components"
)

func fixture() collect.Snapshot {
	return collect.Snapshot{
		At: time.Now(), Hostname: "router", Model: "K2P", Release: "OpenWrt", Arch: "mips",
		Kernel: "4.14", Uptime: 41*time.Hour + 50*time.Minute, Took: 12 * time.Millisecond,
		Mem:    collect.MemInfo{Total: 64 << 20, Available: 32 << 20},
		Ifaces: []collect.Iface{{Name: "br-lan", Up: true, Addrs: []string{"192.168.1.1"}, RxRate: 2048}},
		Procs:  []collect.Proc{{PID: 1, RSS: 1 << 20, Name: "procd", CPU: 5}},
		ProcN:  1,
		Wifi: collect.WifiInfo{Supported: true, Bands: []collect.WifiBand{
			{
				Name: "2.4 GHz", Radio: "radio0", Iface: "default_radio0", Enabled: true,
				SSID: "OpenWrt", Key: "secret123", Security: "psk2", Channel: "6",
			},
			{
				Name: "5 GHz", Radio: "radio1", Iface: "default_radio1", Enabled: true,
				SSID: "OpenWrt-5G", Key: "secret456", Security: "sae", Channel: "36",
			},
		}},
	}
}

func newTestModel() model {
	m := model{
		col: collect.NewCollector(), snap: fixture(), tab: tabSettings, w: 100, h: 24,
		langSel: kit.LangChoice(), langFocus: kit.LangChoice(),
		traffic: network.NewTracker(),
		hist:    status.NewTracker(),
		cron:    cron.NewForm(),
	}
	m.subOpen = settings.Wifi
	m.syncForm()
	return m
}

// Regression: the speed pane sits to the right of the sidebar; a click must
// translate into pane coordinates (sidebar width + gutter), and the sidebar
// itself must be keyboard-navigable like the settings tab.
func TestSpeedPaneMouse(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.tab = 1
	m.netOpen, m.netActive = network.TabSpeed, true

	startY := bandRowY + speedtest.StartRow
	click := func(x, y int) tea.MouseMsg {
		return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
	}

	// A click on the start button starts the measurement flow (the returned
	// command is not executed here, so no network is touched).
	mm, cmd := m.handleMouse(click(kit.SidebarW+10, startY))
	m = mm.(model)
	if cmd == nil {
		t.Fatal("start click produced no command")
	}
	if m.speed.Running {
		t.Fatal("start click set Running without a prepared server")
	}

	// One row below the button must do nothing.
	mm, cmd = m.handleMouse(click(kit.SidebarW+10, startY+1))
	m = mm.(model)
	if cmd != nil {
		t.Fatal("click beside the button started something")
	}
}

// The network sidebar navigates like the settings sidebar: j/k previews,
// enter opens, esc returns.
func TestNetworkSidebarKeyboard(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.tab = 1

	// enter opens the previewed sub-tab (overview by default)
	mm, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if !m.netActive || m.netOpen != network.TabOverview {
		t.Fatalf("enter: active=%v open=%d", m.netActive, m.netOpen)
	}

	// esc returns to the sidebar
	mm, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = mm.(model)
	if m.netActive {
		t.Fatal("esc did not return to the sidebar")
	}

	// j moves to the speed test and previews it; enter opens it
	mm, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	m = mm.(model)
	if m.netSel != network.TabSpeed || m.netOpen != network.TabSpeed {
		t.Fatalf("j: sel=%d open=%d", m.netSel, m.netOpen)
	}
	mm, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if !m.netActive || m.netOpen != network.TabSpeed {
		t.Fatalf("enter speed: active=%v open=%d", m.netActive, m.netOpen)
	}

	// inside the speed pane, enter starts the flow (prepare command; not run)
	mm, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if cmd == nil {
		t.Fatal("enter in speed pane did not start the flow")
	}
	if m.speed.Running {
		t.Fatal("running set before the server was prepared")
	}

	// esc leaves the pane instead of quitting
	mm, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyEsc})
	m = mm.(model)
	if m.netActive {
		t.Fatal("esc did not leave the speed pane")
	}
}

func TestTabMarkersWithoutColor(t *testing.T) {
	kit.SetLang("en")
	kit.SetColor(false)
	t.Cleanup(func() { kit.SetColor(true) })

	m := newTestModel()
	m.tab = 1
	bar := m.tabBar()
	if !strings.Contains(bar, "[Network]") {
		t.Fatalf("marker missing: %q", bar)
	}
	hits := m.tabHits()
	col := lipgloss.Width(bar[:strings.Index(bar, "[Network]")]) + 1
	if hits[1][0] != col {
		t.Fatalf("hit start %d, rendered %d", hits[1][0], col)
	}
}

func TestSettingsEnterAndBack(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.tab = 0
	m.setTab(tabSettings)
	if m.subActive || m.subOpen != settings.Wifi || m.subSel != 0 {
		t.Fatalf("default section not preselected: active=%v open=%v sel=%d", m.subActive, m.subOpen, m.subSel)
	}

	// Enter opens the highlighted section, Esc goes back to the sidebar.
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if !m.subActive {
		t.Fatal("enter did not open the section")
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if mm.(model).subActive {
		t.Fatal("esc did not return to the sidebar")
	}
}

func TestClickToggleAndInput(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.subActive = true
	off := wifi.FormRowOffset(m.settingsState().Wifi)

	click := func(m model, row int) model {
		mm, _ := m.handleMouse(tea.MouseMsg{
			Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
			X: 40, Y: 4 + row,
		})
		return mm.(model)
	}

	got := click(m, off+wifi.FocusEnabled)
	if got.form.Enabled.On || got.form.Focus != wifi.FocusEnabled {
		t.Fatalf("enabled click: enabled=%v focus=%d", got.form.Enabled.On, got.form.Focus)
	}
	got = click(m, off+wifi.FocusSSID)
	if !got.form.Editing || got.form.Focus != wifi.FocusSSID || got.form.SSID.Value != "OpenWrt" {
		t.Fatal("ssid click did not start editing")
	}
	// cycling rows only take focus
	got = click(m, off+wifi.FocusChannel)
	if got.form.Channel.Value() != "6" || got.form.Focus != wifi.FocusChannel {
		t.Fatal("channel click changed the value")
	}
}

func TestButtonsShareRow(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.subActive = true
	m.form.SSID.Value = "new"
	off := wifi.FormRowOffset(m.settingsState().Wifi)
	y := 4 + off + 6

	body := strings.Join(m.body(), "\n")
	if !strings.Contains(body, "[ Apply ]") || !strings.Contains(body, "[ Cancel ]") {
		t.Fatal("buttons missing")
	}
	if strings.Count(body, "[ Apply ]") != 1 {
		t.Fatal("unexpected button count")
	}

	btns := components.Buttons{Labels: [2]string{kit.Tr.Apply, kit.Tr.Cancel}}
	hits := btns.Hits()
	mm, cmd := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: kit.SidebarW + 2 + hits[0][0], Y: y})
	if got := mm.(model); got.form.Focus != wifi.FocusApply || !got.form.Busy || cmd == nil {
		t.Fatal("apply click did not dispatch")
	}
	mm, _ = m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: kit.SidebarW + 2 + hits[1][0], Y: y})
	if got := mm.(model); got.form.SSID.Value != "OpenWrt" {
		t.Fatal("cancel click did not revert")
	}
}

func TestStatusBarTiming(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.tab = 0
	m.refresh = time.Second
	v := m.View()
	if !strings.Contains(v, "Uptime") || !strings.Contains(v, "refresh") {
		t.Fatalf("status bar missing timing info:\n%s", v)
	}
	if !strings.Contains(v, "12 ms") {
		t.Fatalf("sample duration missing:\n%s", v)
	}
}

// dirtySettings leaves every piece of Settings UI state away from its default.
func dirtySettings(m model) model {
	m.subOpen = settings.Language
	m.subSel = 1
	m.subActive = true
	m.band = 1
	m.form = wifi.FromBand(m.snap.Wifi.Bands[1], wifi.FocusSecurity)
	m.form.SSID.Value = "dirty"
	m.form.Editing = true
	m.form.Busy = false
	m.formNote, m.formErr = "note", "err"
	m.langSel, m.langFocus = 2, 2
	return m
}

func checkReset(t *testing.T, m model, how string) {
	t.Helper()
	if m.subOpen != settings.Wifi || m.subSel != 0 || m.subActive {
		t.Fatalf("%s: section open=%v sel=%d active=%v", how, m.subOpen, m.subSel, m.subActive)
	}
	if m.band != 0 {
		t.Fatalf("%s: band=%d", how, m.band)
	}
	if m.form.Focus != wifi.FocusEnabled || m.form.Editing || m.form.Changed() {
		t.Fatalf("%s: form focus=%d editing=%v changed=%v",
			how, m.form.Focus, m.form.Editing, m.form.Changed())
	}
	if m.form.SSID.Value != "OpenWrt" || m.form.Key.Value != "secret123" || m.form.Channel.Value() != "6" {
		t.Fatalf("%s: form reloaded wrong values: %+v", how, m.form)
	}
	if m.formNote != "" || m.formErr != "" {
		t.Fatalf("%s: messages survived: %q %q", how, m.formNote, m.formErr)
	}
	if m.langFocus != m.langSel || m.langSel != kit.LangChoice() {
		t.Fatalf("%s: language focus=%d sel=%d", how, m.langFocus, m.langSel)
	}
}

func TestSettingsResetOnEnter(t *testing.T) {
	kit.SetLang("en")
	t.Cleanup(func() { kit.SetLang("en") })

	// direct setTab
	m := dirtySettings(newTestModel())
	m.tab = 0
	m.setTab(tabSettings)
	checkReset(t, m, "setTab")

	// digit key 4 from another tab
	m = dirtySettings(newTestModel())
	m.tab = 0
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	checkReset(t, mm.(model), "key 4")

	// tab / shift+tab cycling into Settings
	m = dirtySettings(newTestModel())
	m.tab = 2
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	checkReset(t, mm.(model), "tab")

	m = dirtySettings(newTestModel())
	m.tab = 0
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	checkReset(t, mm.(model), "shift+tab")

	// clicking the Settings tab again while already inside it
	m = dirtySettings(newTestModel())
	m.tab = tabSettings
	hits := m.tabHits()
	x := (hits[tabSettings][0] + hits[tabSettings][1]) / 2
	mm, _ = m.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, X: x, Y: tabRowY,
	})
	checkReset(t, mm.(model), "tab click")
}

func TestBandArrows(t *testing.T) {
	kit.SetLang("en")
	t.Cleanup(func() { kit.SetLang("en") })

	// preselected: the sidebar only previews the pane, so the arrows still
	// belong to the main tabs instead of the band selector
	m := newTestModel()
	m.tab = 0
	m.setTab(tabSettings)
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = mm.(model)
	if m.tab != 0 || m.band != 0 || m.subActive {
		t.Fatalf("preview right should switch tabs: tab=%d band=%d active=%v", m.tab, m.band, m.subActive)
	}

	// back in Settings and inside the pane, the same keys switch bands
	m = newTestModel()
	m.tab = 0
	m.setTab(tabSettings)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = mm.(model)
	if m.band != 1 || !m.subActive {
		t.Fatalf("active right: band=%d active=%v", m.band, m.subActive)
	}
	if m.form.SSID.Value != "OpenWrt-5G" {
		t.Fatalf("form not reloaded for band 1: %q", m.form.SSID.Value)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = mm.(model)
	if m.band != 0 || m.form.SSID.Value != "OpenWrt" {
		t.Fatalf("active left: band=%d ssid=%q", m.band, m.form.SSID.Value)
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = mm.(model)
	if m.band != 1 || !m.subActive {
		t.Fatalf("active right: band=%d active=%v", m.band, m.subActive)
	}
	// a focused value row still switches bands until Enter enters it
	m.form.Focus = wifi.FocusChannel
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = mm.(model)
	if m.band != 0 || m.form.Editing {
		t.Fatalf("focused channel row: band=%d editing=%v", m.band, m.form.Editing)
	}

	// Enter turns the row into an editor; arrows now change the value
	ch := m.form.Channel.Value()
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mm.(model)
	if !m.form.Editing {
		t.Fatal("enter did not open the channel row")
	}
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = mm.(model)
	if m.band != 0 || m.form.Channel.Value() == ch {
		t.Fatalf("channel step: band=%d channel=%q", m.band, m.form.Channel.Value())
	}
	// Esc restores the value and leaves the row
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = mm.(model)
	if m.form.Editing || m.form.Channel.Value() != ch {
		t.Fatalf("esc: editing=%v channel=%q", m.form.Editing, m.form.Channel.Value())
	}
	// a dirty form blocks the band change instead of dropping the edit
	m.form.SSID.Value = "dirty"
	m.form.Focus = wifi.FocusSSID
	band := m.band
	mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = mm.(model)
	if m.band != band || m.formNote != kit.Tr.ApplyFirst {
		t.Fatalf("dirty block: band=%d note=%q", m.band, m.formNote)
	}
}

func TestListRowsNeedEnter(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.tab = 0
	m.setTab(tabSettings)
	mm, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // open Wi-Fi
	m = mm.(model)

	for _, focus := range []int{wifi.FocusChannel, wifi.FocusSecurity} {
		// start every case from a clean form so the band switch is allowed
		m.form = wifi.FromBand(m.snap.Wifi.Bands[m.band], focus)
		before := m.form
		band := m.band

		// arrows on a merely focused row switch the band, not the value
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = mm.(model)
		if m.form.Editing {
			t.Fatalf("focus %d: arrows opened the row", focus)
		}
		if m.band == band {
			t.Fatalf("focus %d: arrows should have switched the band", focus)
		}
		// the form now shows the other band, untouched
		shown := m.snap.Wifi.Bands[m.band]
		if m.form.Channel.Value() != shown.Channel || m.form.Security.Value() != shown.Security {
			t.Fatalf("focus %d: form out of sync after band switch: %+v", focus, m.form)
		}
		m.form.Focus = focus
		before = m.form

		// Enter opens the row; arrows now step the value
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = mm.(model)
		if !m.form.Editing {
			t.Fatalf("focus %d: enter did not open the row", focus)
		}
		band = m.band
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = mm.(model)
		if m.band != band {
			t.Fatalf("focus %d: open row should not switch bands", focus)
		}
		if focus == wifi.FocusChannel && m.form.Channel.Value() == before.Channel.Value() {
			t.Fatal("channel did not step")
		}
		if focus == wifi.FocusSecurity && m.form.Security.Value() == before.Security.Value() {
			t.Fatal("security did not step")
		}

		// Enter commits and closes the row
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = mm.(model)
		if m.form.Editing {
			t.Fatalf("focus %d: enter did not close the row", focus)
		}
		// Esc inside the row restores the pre-edit value
		m.form.Focus = focus
		want := m.form.FocusValue()
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		m = mm.(model)
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
		m = mm.(model)
		mm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		m = mm.(model)
		if m.form.Editing || m.form.FocusValue() != want {
			t.Fatalf("focus %d: esc did not restore %q (got %q)", focus, want, m.form.FocusValue())
		}
	}
}

func TestClickListRowOnlyFocuses(t *testing.T) {
	kit.SetLang("en")
	m := newTestModel()
	m.subActive = true
	off := wifi.FormRowOffset(m.settingsState().Wifi)
	before := m.form
	mm, _ := m.handleMouse(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress,
		X: 40, Y: 4 + off + wifi.FocusChannel})
	got := mm.(model)
	if got.form.Editing || got.form.Focus != wifi.FocusChannel ||
		got.form.Channel.Value() != before.Channel.Value() {
		t.Fatalf("channel click: editing=%v focus=%d channel=%q",
			got.form.Editing, got.form.Focus, got.form.Channel.Value())
	}
}

func TestBootTimeInTabBar(t *testing.T) {
	kit.SetLang("en")
	kit.SetColor(false)
	t.Cleanup(func() { kit.SetColor(true) })

	m := newTestModel()
	m.boot = 12 * time.Millisecond

	bar := m.topBar()
	if !strings.Contains(bar, "startup 12 ms") {
		t.Fatalf("boot label missing: %q", bar)
	}
	if w := lipgloss.Width(bar); w != m.w {
		t.Fatalf("tab bar width %d, want %d", w, m.w)
	}
	// the boot label sits at the right edge
	if !strings.HasSuffix(strings.TrimRight(bar, " "), "startup 12 ms") {
		t.Fatalf("boot label not right aligned: %q", bar)
	}

	// a narrow terminal drops the label instead of clipping the tabs
	m.w = 30
	if strings.Contains(m.topBar(), "startup") {
		t.Fatal("boot label kept on a narrow bar")
	}
}
