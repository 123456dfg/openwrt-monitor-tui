package status

import (
	"openwrt-monitor-tui/ui/i18n"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/style"
)

func snap(mem collect.MemInfo) collect.Snapshot {
	return collect.Snapshot{At: time.Now(), Mem: mem, CPUUsage: 20}
}

func TestSwapRowHiddenWithoutSwap(t *testing.T) {
	i18n.SetLang("en")
	style.SetColor(false)
	t.Cleanup(func() { style.SetColor(true) })

	body := strings.Join(Body(snap(collect.MemInfo{Total: 100, Available: 40}), nil), "\n")
	if strings.Contains(body, style.Tr.Swap) {
		t.Fatalf("swap row shown without swap:\n%s", body)
	}

	withSwap := snap(collect.MemInfo{Total: 100, Available: 40, SwapTotal: 100, SwapFree: 70})
	body = strings.Join(Body(withSwap, nil), "\n")
	if !strings.Contains(body, style.Tr.Swap) {
		t.Fatalf("swap row missing:\n%s", body)
	}
}

func TestTrackerCharts(t *testing.T) {
	i18n.SetLang("en")
	style.SetColor(false)
	t.Cleanup(func() { style.SetColor(true) })

	tr := NewTracker()
	for i := 0; i < 20; i++ {
		tr.Add(snap(collect.MemInfo{Total: 100, Available: uint64(100 - i)}))
	}
	cpu, mem := tr.CPUChart(), tr.MemChart()
	if cpu == "" || mem == "" {
		t.Fatalf("charts missing: %q %q", cpu, mem)
	}
	if lipgloss.Width(cpu) != chartW || lipgloss.Width(mem) != chartW {
		t.Fatalf("chart width: %d %d", lipgloss.Width(cpu), lipgloss.Width(mem))
	}
	if tr.SwapChart() != "" {
		t.Fatalf("swap chart without swap: %q", tr.SwapChart())
	}

	// once swap exists its history starts
	tr.Add(snap(collect.MemInfo{Total: 100, Available: 50, SwapTotal: 100, SwapFree: 90}))
	if tr.SwapChart() == "" {
		t.Fatal("swap chart missing after swap appeared")
	}

	// the curves are drawn next to the meters
	body := strings.Join(Body(snap(collect.MemInfo{Total: 100, Available: 40}), tr), "\n")
	if !strings.Contains(body, cpu) {
		t.Fatalf("cpu curve not rendered:\n%s", body)
	}
}

func TestHistoryIsCapped(t *testing.T) {
	tr := NewTracker()
	for i := 0; i < chartW*3; i++ {
		tr.Add(snap(collect.MemInfo{Total: 100, Available: 50}))
	}
	if len(tr.cpu) != chartW || len(tr.mem) != chartW {
		t.Fatalf("history length: %d %d", len(tr.cpu), len(tr.mem))
	}
}
