package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/ui/network/portscan"
	speed "openwrt-monitor-tui/ui/network/speedtest"
)

// updateSpeed handles the keyboard inside the entered speed test pane; Esc
// back to the sidebar is handled by handleKey. Global keys (q, tab, r, …)
// are consumed there too.
func (m model) updateSpeed(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		return m.startSpeed()
	}
	return m, nil
}

// startSpeed resolves the closest server, then measures download and upload
// through the speedtest-go library, each in its own background command.
func (m model) startSpeed() (tea.Model, tea.Cmd) {
	if m.speed.Running {
		return m, nil
	}
	return m, m.speedPrepareCmd()
}

func (m model) speedPrepareCmd() tea.Cmd {
	prov := speed.Providers[0]
	return func() tea.Msg {
		h, info, err := prov.Native.Prepare()
		return speedPreparedMsg{native: h, server: info, err: err}
	}
}

func (m model) speedDownCmd() tea.Cmd {
	live, h := m.speed.Live, m.speed.Native
	return func() tea.Msg {
		rate, err := speed.Providers[0].Native.Down(h, live)
		return speedDownMsg{rate: rate, err: err}
	}
}

func (m model) speedUpCmd() tea.Cmd {
	live, h := m.speed.Live, m.speed.Native
	return func() tea.Msg {
		rate, err := speed.Providers[0].Native.Up(h, live)
		return speedUpMsg{rate: rate, err: err}
	}
}

// updatePorts handles the keyboard inside the entered port scan pane; Esc
// back to the sidebar is handled by handleKey.
func (m model) updatePorts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ":
		return m.startScan()
	}
	return m, nil
}

// startScan kicks off a background scan through the task package; repeated
// presses while one is running do nothing.
func (m model) startScan() (tea.Model, tea.Cmd) {
	if m.scan.Phase == portscan.PhaseRunning {
		return m, nil
	}
	job := m.scanJob
	m.scan.Phase = portscan.PhaseRunning
	m.scan.Progress = job.Progress()
	m.scan.Err = ""
	return m, func() tea.Msg {
		job.Start() // runs the scan to completion on this goroutine
		res, _ := job.Result()
		return scanDoneMsg{res: res}
	}
}
