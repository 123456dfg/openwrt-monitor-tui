package procs

import (
	"fmt"
	"slices"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
)

// Sort picks which column the process table is ranked by.
type Sort uint8

const (
	ByMem Sort = iota
	ByCPU
)

// DisplayLimit caps how many rows the tab shows; the snapshot carries every
// process, the table shows the top of the current ordering.
const DisplayLimit = 12

// Header layout constants, shared with the mouse layer: the header is body
// line HeaderRow, and the CPU/RSS labels span these x cells.
const HeaderRow = 1

var (
	cpuHit = [2]int{10, 15}
	memHit = [2]int{17, 26}
)

// HeaderClick maps a press on the table header to the column to rank by.
func HeaderClick(x int) (Sort, bool) {
	switch {
	case x >= cpuHit[0] && x <= cpuHit[1]:
		return ByCPU, true
	case x >= memHit[0] && x <= memHit[1]:
		return ByMem, true
	}
	return 0, false
}

// Top ranks procs by the chosen column, heaviest first unless desc is false,
// and returns up to n rows in display order. It sorts a copy — the snapshot
// slice is shared with the collector, so it must not be reordered in place.
//
// This runs once per snapshot (or sort-key change), not per frame: View is
// called on every keypress and scroll, and on a low-end MIPS core even a
// few hundred reflect-based swaps per render would eat into the budget.
// slices.SortFunc compares directly, no reflection.
func Top(procs []collect.Proc, by Sort, desc bool, n int) []collect.Proc {
	if len(procs) == 0 {
		return nil
	}
	sorted := make([]collect.Proc, len(procs))
	copy(sorted, procs)
	cmp := func(a, b collect.Proc) int {
		if by == ByCPU {
			return cmpFloat(a.CPU, b.CPU)
		}
		switch {
		case a.RSS < b.RSS:
			return -1
		case a.RSS > b.RSS:
			return 1
		}
		return 0
	}
	if desc {
		old := cmp
		cmp = func(a, b collect.Proc) int { return old(b, a) }
	}
	slices.SortFunc(sorted, cmp)
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// Body renders the process table from rows that Top already ranked and
// capped, so rendering never pays for sorting. by/desc only drive the header
// arrows.
func Body(rows []collect.Proc, total int, by Sort, desc bool) []string {
	head := kit.PadRight(kit.Tr.PID, 7) + " " + kit.PadLeft(mark(kit.Tr.CPUPct, by == ByCPU, desc), 6) + " " +
		kit.PadLeft(mark(kit.Tr.RSS, by == ByMem, desc), 10) + "   " + kit.Tr.Command
	out := []string{"", "  " + kit.Label.Render(head)}
	if len(rows) == 0 {
		out = append(out, "  "+kit.Dim.Render(kit.Tr.NoProcs))
	}
	for _, p := range rows {
		out = append(out, fmt.Sprintf("  %-7d %5.1f%% %10s   %s",
			p.PID, p.CPU, kit.HumanBytes(p.RSS), p.Name))
	}
	out = append(out, "", "  "+kit.Dim.Render(fmt.Sprintf(kit.Tr.ProcSummary, total, len(rows))))
	return out
}

// mark appends the direction arrow to a header label when that column is the
// active sort. The arrow fits inside the existing column width, so the row
// layout never shifts.
func mark(label string, active, desc bool) string {
	if !active {
		return label
	}
	if kit.ASCII() {
		if desc {
			return label + "v"
		}
		return label + "^"
	}
	if desc {
		return label + "▼"
	}
	return label + "▲"
}
