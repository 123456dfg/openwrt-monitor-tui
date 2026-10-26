package network

import (
	"fmt"
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/network/portscan"
	"openwrt-monitor-tui/ui/network/speedtest"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets"
)

// Sub-tabs of the Network pane, shown in a settings-style sidebar.
const (
	TabOverview = iota
	TabSpeed
	TabPorts
)

const maxAddrW = 24

// Body merges the sidebar (overview / speed test / ports) and the pane of the
// open sub-tab, mirroring the settings tab layout: sidebar rows padded to
// SidebarW, two-cell gutter, then the pane content. The sidebar marks three
// states: the entered sub-tab (solid triangle), the preselected one (dim dot,
// not entered yet) and the untouched row (no marker).
func Body(s collect.Snapshot, hist *Tracker, width, sel, open int, active bool, speed speedtest.State, ports portscan.State) []string {
	if sel < TabOverview || sel > TabPorts {
		sel = TabOverview
	}
	if open < TabOverview || open > TabPorts {
		open = TabOverview
	}

	names := []string{style.Tr.NetTabOverview, style.Tr.NetTabSpeed, style.Tr.NetTabPorts}
	side := make([]string, 0, len(names))
	for i, n := range names {
		switch {
		case i == open && active:
			side = append(side, " "+style.Accent.Render(style.Glyph.Sel+" "+n))
		case i == sel:
			side = append(side, " "+style.Dim.Render(style.Glyph.Dot+" "+n))
		default:
			side = append(side, "  "+style.Label.Render(n))
		}
	}

	var content []string
	switch open {
	case TabSpeed:
		content = speedtest.Pane(speed)
	case TabPorts:
		content = portscan.Pane(ports)
	default:
		content = overview(s, hist, width)
	}

	out := make([]string, max(len(side), len(content)))
	for i := range out {
		var left, right string
		if i < len(side) {
			left = side[i]
		}
		if i < len(content) {
			right = content[i]
		}
		out[i] = strings.TrimRight(style.PadRight(left, style.SidebarW)+"  "+right, " ")
	}
	return out
}

// overview is the interface table and the traffic totals.
func overview(s collect.Snapshot, hist *Tracker, width int) []string {
	addrW := min(maxAddrW, max(10, width-53-chartGap))
	head := style.PadRight(style.Tr.Iface, 10) + " " + style.PadRight(style.Tr.State, 5) + " " +
		style.PadRight(style.Tr.Address, addrW) + " " + style.PadLeft(style.Tr.RxRate, 10) + " " +
		style.PadLeft(style.Tr.TxRate, 10) + strings.Repeat(" ", chartGap) +
		style.PadLeft(style.Tr.Traffic, chartW)
	out := []string{"  " + style.Label.Render(head)}
	if len(s.Ifaces) == 0 {
		out = append(out, "  "+style.Dim.Render(style.Tr.NoIfaces))
	}
	for i, in := range s.Ifaces {
		if i > 0 {
			// One blank line between entries: half of a one-line entry,
			// rounded up to the smallest gap a terminal can draw.
			out = append(out, "")
		}
		state := style.Tr.Down
		if in.Up {
			state = style.Tr.Up
		}
		out = append(out, "  "+style.PadRight(in.Name, 10)+" "+style.PadRight(state, 5)+" "+
			cell(preferredAddr(in.Addrs), addrW)+" "+
			style.PadLeft(rateStr(in.RxRate), 10)+" "+style.PadLeft(rateStr(in.TxRate), 10)+
			strings.Repeat(" ", chartGap)+hist.Chart(in.Name))
	}
	out = append(out, "")
	out = append(out, "  "+style.Dim.Render(fmt.Sprintf(style.Tr.TotalTraffic,
		style.HumanBytes(s.TotalRx()), style.HumanBytes(s.TotalTx()))))
	return out
}

func rateStr(v float64) string {
	if v <= 0 {
		return "-"
	}
	return style.HumanBytes(uint64(v)) + "/s"
}

// preferredAddr picks a stable address to show: global IPv4, then anything not
// link-local.
func preferredAddr(addrs []string) string {
	if len(addrs) == 0 {
		return "-"
	}
	for _, a := range addrs {
		if strings.Contains(a, ".") && !strings.HasPrefix(a, "169.254.") {
			return a
		}
	}
	for _, a := range addrs {
		if !strings.HasPrefix(a, "fe80") {
			return a
		}
	}
	return addrs[0]
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

// cell truncates and pads to n runes so columns line up.
func cell(s string, n int) string {
	s = trunc(s, n)
	if pad := n - len([]rune(s)); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

const (
	// chartW is how many samples the per-interface sparkline keeps.
	chartW = 12
	// chartGap separates the TX rate column from the chart.
	chartGap = 3
)

// Tracker keeps the combined rx+tx rate history of every interface. It is
// fed only while the Network tab is visible, matching the lazy sampling.
type Tracker struct {
	series map[string][]float64
}

func NewTracker() *Tracker {
	return &Tracker{series: map[string][]float64{}}
}

func (t *Tracker) Add(ifaces []collect.Iface) {
	if t == nil {
		return
	}
	for _, in := range ifaces {
		s := append(t.series[in.Name], in.RxRate+in.TxRate)
		if len(s) > chartW {
			s = s[len(s)-chartW:]
		}
		t.series[in.Name] = s
	}
}

// Chart draws the rate history as a curve (see widgets.Chart).
func (t *Tracker) Chart(name string) string {
	var series []float64
	if t != nil {
		series = t.series[name]
	}
	return widgets.Chart(series, chartW)
}
