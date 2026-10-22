package network

import (
	"fmt"
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/widgets"
)

const maxAddrW = 24

// Body renders the Network tab from a snapshot and its rate history.
func Body(s collect.Snapshot, hist *Tracker, width int) []string {
	addrW := min(maxAddrW, max(10, width-53-chartGap))
	head := kit.PadRight(kit.Tr.Iface, 10) + " " + kit.PadRight(kit.Tr.State, 5) + " " +
		kit.PadRight(kit.Tr.Address, addrW) + " " + kit.PadLeft(kit.Tr.RxRate, 10) + " " +
		kit.PadLeft(kit.Tr.TxRate, 10) + strings.Repeat(" ", chartGap) +
		kit.PadLeft(kit.Tr.Traffic, chartW)
	out := []string{"", "  " + kit.Label.Render(head)}
	if len(s.Ifaces) == 0 {
		out = append(out, "  "+kit.Dim.Render(kit.Tr.NoIfaces))
	}
	for i, in := range s.Ifaces {
		if i > 0 {
			// One blank line between entries: half of a one-line entry,
			// rounded up to the smallest gap a terminal can draw.
			out = append(out, "")
		}
		state := kit.Tr.Down
		if in.Up {
			state = kit.Tr.Up
		}
		out = append(out, "  "+kit.PadRight(in.Name, 10)+" "+kit.PadRight(state, 5)+" "+
			cell(preferredAddr(in.Addrs), addrW)+" "+
			kit.PadLeft(rateStr(in.RxRate), 10)+" "+kit.PadLeft(rateStr(in.TxRate), 10)+
			strings.Repeat(" ", chartGap)+hist.Chart(in.Name))
	}
	out = append(out, "")
	out = append(out, "  "+kit.Dim.Render(fmt.Sprintf(kit.Tr.TotalTraffic,
		kit.HumanBytes(s.TotalRx()), kit.HumanBytes(s.TotalTx()))))
	return out
}

func rateStr(v float64) string {
	if v <= 0 {
		return "-"
	}
	return kit.HumanBytes(uint64(v)) + "/s"
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
