// Pane is the port scan tab content: a start action, the live probe counter
// and the per-host results. Row indices here are shared with the mouse layer.
package portscan

import (
	"fmt"
	"strings"

	"openwrt-monitor-tui/task"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// Pane row indices, relative to the first line of Pane's output — which the
// Network tab aligns with the top of its sidebar. The mouse layer subtracts
// that offset before comparing.
const StartRow = 0

// Measurement phases, mirrored by the pane.
const (
	PhaseIdle = iota
	PhaseRunning
	PhaseDone
)

// State is everything the pane shows.
type State struct {
	Phase    int                // PhaseIdle / PhaseRunning / PhaseDone
	Active   bool               // the ports sub-tab has been entered
	Progress *task.ScanProgress // live probe counters, nil when idle
	Hosts    []task.HostResult  // finished scan results
	Err      string
}

// New builds the initial pane state.
func New() State { return State{} }

// Pane renders the port scan content.
func Pane(st State) []string {
	out := make([]string, 0, 8+len(st.Hosts))
	out = append(out, startRow(st), "")
	switch {
	case st.Phase == PhaseRunning && st.Progress != nil:
		out = append(out, "  "+style.Dim.Render(fmt.Sprintf("%s %d/%d",
			style.Tr.PortProbed, st.Progress.Probed.Load(), st.Progress.Total.Load())))
	case st.Err != "":
		out = append(out, "  "+style.Bad.Render(st.Err))
	case st.Phase == PhaseDone:
		open := 0
		for _, h := range st.Hosts {
			open += len(h.Ports)
		}
		out = append(out, "  "+style.Label.Render(fmt.Sprintf(style.Tr.PortSummary,
			len(st.Hosts), open)))
		if len(st.Hosts) == 0 {
			out = append(out, "  "+style.Dim.Render(style.Tr.PortNone))
		} else {
			out = append(out, "")
			out = append(out, hostLines(st.Hosts)...)
		}
	}
	return out
}

// startRow renders the action row. Once the sub-tab has been entered the
// bracketed button turns into the play-style row, matching the speed test.
func startRow(st State) string {
	if st.Active {
		label := style.Tr.PortStart
		row := style.Accent
		if st.Phase == PhaseRunning {
			label, row = style.Tr.PortRunning, style.Dim
		}
		return "  " + row.Render(style.Glyph.Sel+" "+label)
	}
	b := components.Buttons{Labels: [2]string{style.Tr.PortStart, ""}}
	b.Focus = -1
	if st.Phase == PhaseRunning {
		b.Busy, b.BusyLabel = true, style.Tr.PortRunning
	}
	return "  " + b.View()
}

// hostLines renders one row per discovered host: address, MAC and the ports
// that answered; hosts without open ports show a dash.
func hostLines(hosts []task.HostResult) []string {
	out := make([]string, 0, len(hosts))
	for _, h := range hosts {
		ports := "-"
		if len(h.Ports) > 0 {
			p := make([]string, 0, len(h.Ports))
			for _, n := range h.Ports {
				p = append(p, fmt.Sprintf("%d", n))
			}
			ports = strings.Join(p, " ")
		}
		out = append(out, fmt.Sprintf("  %-15s %-17s %s", h.IP, h.MAC, ports))
	}
	return out
}
