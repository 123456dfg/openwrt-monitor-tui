package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/style"
)

// Buttons is a row of one or two actions, focusable with the keyboard and
// clickable in the same layout. Focus is the index of the focused action, or
// -1. An empty second label hides that button, which is how single-action
// rows are built.
type Buttons struct {
	Labels    [2]string
	Focus     int
	Busy      bool
	BusyLabel string // replaces Labels[0] while Busy
}

func (b Buttons) label(i int) string {
	if i == 0 && b.Busy && b.BusyLabel != "" {
		return b.BusyLabel
	}
	return b.Labels[i]
}

// live returns the indices of the buttons that are rendered.
func (b Buttons) live() []int {
	var out []int
	for i := 0; i < 2; i++ {
		if b.label(i) != "" {
			out = append(out, i)
		}
	}
	return out
}

// View renders the actions on one line. The focus marker sits in a fixed
// two-cell prefix so the buttons do not shift when the focus moves.
func (b Buttons) View() string {
	var out strings.Builder
	out.WriteString("  ")
	first := true
	for _, i := range b.live() {
		if !first {
			out.WriteString("   ")
		}
		first = false
		label := "[ " + b.label(i) + " ]"
		if i == b.Focus {
			out.WriteString(style.Glyph.Sel + " " + style.Accent.Render(label))
			continue
		}
		out.WriteString("  " + style.Label.Render(label))
	}
	return out.String()
}

// Hits returns the clickable x range of each rendered button, relative to the
// start of the row (the same origin View uses). Each range includes 10% of
// the button width as padding on both sides, so a click just outside the
// label still lands on it.
func (b Buttons) Hits() [][2]int {
	var hits [][2]int
	x := 2 // View's leading indent
	for _, i := range b.live() {
		w := 2 + lipgloss.Width("[ "+b.label(i)+" ]")
		p := max(1, w/10)
		hits = append(hits, [2]int{x - p, x + w - 1 + p})
		x += w + 3
	}
	return hits
}

// Click handles a press at x (same origin as Hits) by moving Focus to the
// pressed button. It reports whether a button was hit; what pressing it
// means is the pane's business.
func (b *Buttons) Click(x, y int) bool {
	live := b.live()
	for i, h := range b.Hits() {
		if x >= h[0] && x <= h[1] {
			b.Focus = live[i]
			return true
		}
	}
	return false
}
