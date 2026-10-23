package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/kit"
)

// Tabs is a horizontal tab strip: the labels of the tabs with one active
// entry. The active tab gets the accent colour, or brackets when colours are
// off — which widens it by two cells, and Hits accounts for that so clicks
// stay aligned either way.
type Tabs struct {
	Names []string
	Cur   int
}

func (t Tabs) clamp(i int) int {
	if i < 0 || i >= len(t.Names) {
		return -1
	}
	return i
}

// View renders the strip with three spaces between tabs.
func (t Tabs) View() string {
	var b strings.Builder
	for i, n := range t.Names {
		if i > 0 {
			b.WriteString("   ")
		}
		if i == t.clamp(t.Cur) {
			if kit.ColorOn() {
				b.WriteString(kit.Accent.Render(n))
			} else {
				b.WriteString("[" + n + "]")
			}
			continue
		}
		b.WriteString(kit.Label.Render(n))
	}
	return b.String()
}

// Hits returns the clickable x range of every tab. x0 is the terminal column
// where the strip content starts, i.e. after whatever indent prefixes View's
// output on the line.
func (t Tabs) Hits(x0 int) [][2]int {
	hits := make([][2]int, 0, len(t.Names))
	for i, n := range t.Names {
		w := lipgloss.Width(n)
		if i == t.clamp(t.Cur) && !kit.ColorOn() {
			w += 2 // the brackets View adds around the active tab
		}
		hits = append(hits, [2]int{x0, x0 + w - 1})
		x0 += w + 3
	}
	return hits
}

// Click maps a press at x (with x0 the strip's content origin, the same
// origin Hits uses) to the tapped tab index; -1 when nothing was hit.
func (t Tabs) Click(x0, x int) int {
	for i, h := range t.Hits(x0) {
		if x >= h[0] && x <= h[1] {
			return i
		}
	}
	return -1
}

// Wheel steps the active tab by d and wraps around, for strips that are
// option lists too (sub-tabs, radio-like rows). It reports whether there was
// more than one tab to move within.
func (t *Tabs) Wheel(d int) bool {
	if len(t.Names) < 2 {
		return false
	}
	t.Cur = ((t.Cur+d)%len(t.Names) + len(t.Names)) % len(t.Names)
	return true
}
