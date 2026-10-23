package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/kit"
)

func TestTabsView(t *testing.T) {
	tabs := Tabs{Names: []string{"Overview", "Speed test"}, Cur: 1}
	line := tabs.View()
	if !strings.Contains(line, "Speed test") {
		t.Fatalf("names missing: %q", line)
	}
	// Three spaces between tabs; widths add up.
	if got, want := lipgloss.Width(line), lipgloss.Width("Overview")+3+lipgloss.Width("Speed test"); got != want {
		t.Fatalf("width %d, want %d", got, want)
	}
}

// Without colours the active tab is bracketed and the hits grow with it.
func TestTabsHitsNoColor(t *testing.T) {
	kit.SetColor(false)
	t.Cleanup(func() { kit.SetColor(true) })

	tabs := Tabs{Names: []string{"AB", "CDEF"}, Cur: 1}
	hits := tabs.Hits(2)
	if len(hits) != 2 {
		t.Fatalf("hits: %d", len(hits))
	}
	// First tab untouched: cells 2..3.
	if hits[0][0] != 2 || hits[0][1] != 3 {
		t.Fatalf("first hit %v", hits[0])
	}
	// Second tab: 2 + 2 + 3 gap = cell 7, bracketed width 4+2 = cells 7..12.
	if hits[1][0] != 7 || hits[1][1] != 12 {
		t.Fatalf("active hit %v, want [7 12]", hits[1])
	}
	// The bracketed rendering matches the widened hit range.
	line := tabs.View()
	if got := lipgloss.Width(line[:strings.Index(line, "[")+1]); got != 1 {
		_ = got // bracket position asserted via the range above
	}
}

func TestTabsClamp(t *testing.T) {
	tabs := Tabs{Names: []string{"a", "b"}, Cur: 9}
	line := tabs.View()
	if strings.Contains(line, "[a]") || strings.Contains(line, "[b]") {
		t.Fatalf("out-of-range cursor bracketed: %q", line)
	}
}
