package status

import (
	"fmt"
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets"
)

// chartW is how many samples each gauge history keeps.
const chartW = 12

// Tracker keeps the last samples of CPU, memory and swap so the Status page can
// draw a curve next to each meter. It is fed only while that tab is visible.
type Tracker struct {
	cpu, mem, swap []float64
}

func NewTracker() *Tracker { return &Tracker{} }

func (t *Tracker) Add(s collect.Snapshot) {
	if t == nil {
		return
	}
	mem := s.Mem
	t.cpu = push(t.cpu, collect.ClampPct(s.CPUUsage))
	t.mem = push(t.mem, pct(memUsed(mem), mem.Total))
	if mem.SwapTotal > 0 {
		t.swap = push(t.swap, pct(swapUsed(mem), mem.SwapTotal))
	}
}

func push(series []float64, v float64) []float64 {
	series = append(series, v)
	if len(series) > chartW {
		series = series[len(series)-chartW:]
	}
	return series
}

func (t *Tracker) CPUChart() string {
	if t == nil {
		return ""
	}
	return t.chart(t.cpu)
}

func (t *Tracker) MemChart() string {
	if t == nil {
		return ""
	}
	return t.chart(t.mem)
}

func (t *Tracker) SwapChart() string {
	if t == nil {
		return ""
	}
	return t.chart(t.swap)
}

func (t *Tracker) chart(series []float64) string {
	if len(series) == 0 {
		return ""
	}
	return widgets.Chart(series, chartW)
}

func memUsed(m collect.MemInfo) uint64 {
	if m.Total > m.Available {
		return m.Total - m.Available
	}
	return 0
}

func swapUsed(m collect.MemInfo) uint64 {
	if m.SwapTotal > m.SwapFree {
		return m.SwapTotal - m.SwapFree
	}
	return 0
}

// Body renders the Status tab from a snapshot and its gauge history.
func Body(s collect.Snapshot, hist *Tracker) []string {
	up := style.FormatUptime(s.Uptime)
	mem := s.Mem
	used := memUsed(mem)

	out := []string{
		"",
		field(style.Tr.Hostname, s.Hostname),
		field(style.Tr.Model, style.Fallback(s.Model, "-")),
		field(style.Tr.Kernel, "Linux "+style.Fallback(s.Kernel, "?")),
		field(style.Tr.Uptime, up),
		field(style.Tr.Load, fmt.Sprintf("%.2f  %.2f  %.2f", s.Load[0], s.Load[1], s.Load[2])),
		"",
		meter(style.Tr.CPU, s.CPUUsage, cpuDetail(s), hist.CPUChart()),
		meter(style.Tr.Memory, pct(used, mem.Total),
			fmt.Sprintf("%s / %s", style.HumanBytes(used), style.HumanBytes(mem.Total)), hist.MemChart()),
	}
	// A device without swap has nothing to show, so the row (and its curve)
	// is omitted instead of printing "not configured".
	if mem.SwapTotal > 0 {
		swap := swapUsed(mem)
		out = append(out, meter(style.Tr.Swap, pct(swap, mem.SwapTotal),
			fmt.Sprintf("%s / %s", style.HumanBytes(swap), style.HumanBytes(mem.SwapTotal)), hist.SwapChart()))
	}
	if s.TempC >= 0 {
		out = append(out, field(style.Tr.Temp, fmt.Sprintf("%.1f °C", s.TempC)))
	}
	out = append(out, "")
	for _, d := range s.Disks {
		used := d.Total - d.Free
		if d.Total < d.Free {
			used = 0
		}
		detail := fmt.Sprintf(style.Tr.FreeOf, style.HumanBytes(d.Free), style.HumanBytes(d.Total))
		if d.FSType != "" {
			detail += "  " + d.FSType
		}
		out = append(out, meter(d.Mount, pct(used, d.Total), detail, ""))
	}
	return out
}

// cpuDetail shows the current clock, falling back to the advertised maximum
// when the driver publishes no live value, and says so when neither exists.
func cpuDetail(s collect.Snapshot) string {
	if f := fmtFreq(s.CPUFreq); f != "" {
		return f
	}
	if f := fmtFreq(s.CPUFreqMax); f != "" {
		return fmt.Sprintf(style.Tr.FreqMax, f)
	}
	return style.Tr.ClockNA
}

func meter(name string, p float64, detail, chart string) string {
	filled := int(p / 100 * style.BarWidth)
	bar := strings.Repeat(style.Glyph.BarFull, filled) +
		strings.Repeat(style.Glyph.BarEmpty, style.BarWidth-filled)
	row := style.Good
	switch {
	case p >= 90:
		row = style.Bad
	case p >= 70:
		row = style.Warn
	}
	line := "  " + style.PadRight(name, 10) + " " + row.Render(bar) +
		fmt.Sprintf(" %4.0f%%", p)
	if chart != "" {
		line += "  " + chart
	}
	if detail != "" {
		line += "   " + style.Dim.Render(detail)
	}
	return line
}

func field(name, value string) string {
	return "  " + style.Label.Render(style.PadRight(name, 10)) + " " + value
}

func pct(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return collect.ClampPct(float64(used) / float64(total) * 100)
}

func fmtFreq(mhz float64) string {
	switch {
	case mhz <= 0:
		return ""
	case mhz >= 1000:
		return fmt.Sprintf("%.2f GHz", mhz/1000)
	default:
		return fmt.Sprintf("%.0f MHz", mhz)
	}
}
