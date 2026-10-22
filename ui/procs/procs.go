package procs

import (
	"fmt"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
)

// Body renders the Processes tab from a snapshot.
func Body(s collect.Snapshot) []string {
	head := kit.PadRight(kit.Tr.PID, 7) + " " + kit.PadLeft(kit.Tr.CPUPct, 6) + " " +
		kit.PadLeft(kit.Tr.RSS, 10) + "   " + kit.Tr.Command
	out := []string{"", "  " + kit.Label.Render(head)}
	if len(s.Procs) == 0 {
		out = append(out, "  "+kit.Dim.Render(kit.Tr.NoProcs))
	}
	for _, p := range s.Procs {
		out = append(out, fmt.Sprintf("  %-7d %5.1f%% %10s   %s",
			p.PID, p.CPU, kit.HumanBytes(p.RSS), p.Name))
	}
	out = append(out, "", "  "+kit.Dim.Render(fmt.Sprintf(kit.Tr.ProcSummary,
		s.ProcN, len(s.Procs))))
	return out
}
