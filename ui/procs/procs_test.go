package procs

import (
	"fmt"
	"strings"
	"testing"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
)

func sample() collect.Snapshot {
	return collect.Snapshot{
		ProcN: 4,
		Procs: []collect.Proc{
			{PID: 1, Name: "low", RSS: 100, CPU: 1},
			{PID: 2, Name: "topmem", RSS: 900, CPU: 5},
			{PID: 3, Name: "topcpu", RSS: 300, CPU: 90},
			{PID: 4, Name: "mid", RSS: 500, CPU: 40},
		},
	}
}

// pids renders the ranked rows and returns their PID column in display order.
func pids(t *testing.T, s collect.Snapshot, by Sort, desc bool) []string {
	t.Helper()
	rows := Top(s.Procs, by, desc, DisplayLimit)
	var out []string
	for _, ln := range Body(rows, s.ProcN, by, desc) {
		f := strings.Fields(ln)
		if len(f) == 0 && len(out) > 0 {
			break // blank line before the summary ends the table
		}
		if len(f) > 0 && f[0] >= "1" && f[0] <= "9" {
			out = append(out, f[0])
		}
	}
	return out
}

func TestSortMemoryDescending(t *testing.T) {
	got := pids(t, sample(), ByMem, true)
	want := []string{"2", "4", "3", "1"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("mem desc: got %v, want %v", got, want)
	}
}

func TestSortCPU(t *testing.T) {
	if got := pids(t, sample(), ByCPU, true); got[0] != "3" {
		t.Fatalf("cpu desc top: got %v", got)
	}
	if got := pids(t, sample(), ByCPU, false); got[0] != "1" {
		t.Fatalf("cpu asc top: got %v", got)
	}
}

// The snapshot slice is shared with the collector and must not be reordered.
func TestTopDoesNotMutateInput(t *testing.T) {
	s := sample()
	Top(s.Procs, ByCPU, true, DisplayLimit)
	if s.Procs[0].PID != 1 {
		t.Fatalf("input slice was reordered: first PID %d", s.Procs[0].PID)
	}
}

func TestHeaderMark(t *testing.T) {
	defer kit.SetUnicode()
	rows := Top(sample().Procs, ByCPU, true, DisplayLimit)
	kit.SetUnicode()
	out := strings.Join(Body(rows, 4, ByCPU, true), "\n")
	if !strings.Contains(out, "CPU%▼") || strings.Contains(out, "RSS▼") {
		t.Fatalf("unicode mark missing: %s", out)
	}
	kit.SetASCII()
	out = strings.Join(Body(rows, 4, ByMem, false), "\n")
	if !strings.Contains(out, "RSS^") || strings.Contains(out, "CPU%v") {
		t.Fatalf("ascii mark missing: %s", out)
	}
}

func TestDisplayLimit(t *testing.T) {
	s := sample()
	for i := range 30 {
		s.Procs = append(s.Procs, collect.Proc{PID: 100 + i, Name: fmt.Sprintf("p%d", i), RSS: 10})
	}
	s.ProcN = len(s.Procs)
	rows := Top(s.Procs, ByMem, true, DisplayLimit)
	if len(rows) != DisplayLimit {
		t.Fatalf("Top returned %d rows, limit is %d", len(rows), DisplayLimit)
	}
}
