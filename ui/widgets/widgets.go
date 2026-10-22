// Package widgets contains the small form controls the settings pages share:
// a checkbox, a radio list, a single-line text input and a select. They own
// only their value and drawing; the caller decides which one has the focus and
// routes keys to it.
package widgets

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/ui/kit"
)

// Checkbox is a boolean switch rendered as [x] / [ ].
type Checkbox struct {
	On bool
}

func (c *Checkbox) Toggle() { c.On = !c.On }

func (c Checkbox) String() string {
	if c.On {
		return "[x]"
	}
	return "[ ]"
}

// Radio is a single-choice list. Cursor is the highlighted row and Sel is the
// confirmed choice, so moving the cursor previews without applying anything.
type Radio struct {
	Options []string
	Cursor  int
	Sel     int
}

func NewRadio(sel int) Radio { return Radio{Cursor: sel, Sel: sel} }

// Move shifts the cursor by d inside the list and wraps around.
func (r *Radio) Move(d int) {
	if len(r.Options) == 0 {
		return
	}
	r.Cursor = ((r.Cursor+d)%len(r.Options) + len(r.Options)) % len(r.Options)
}

// Choose confirms the highlighted row.
func (r *Radio) Choose() { r.Sel = r.Cursor }

// Lines renders one line per option: a cursor marker, a (o)/( ) box and the
// label. A negative Cursor hides the marker, which is what a previewed but not
// entered pane wants.
func (r Radio) Lines(labels []string) []string {
	out := make([]string, 0, len(labels))
	for i, label := range labels {
		marker, style := "  ", kit.Label
		if i == r.Cursor {
			marker, style = kit.Glyph.Sel+" ", kit.Accent
		}
		box := "( )"
		if i == r.Sel {
			box = "(o)"
		}
		out = append(out, marker+style.Render(box+" "+label))
	}
	return out
}

// Input is a single-line text field. Editing is driven by the caller: only one
// row can be edited at a time, so that flag stays with the form.
type Input struct {
	Value string
	Limit int // max runes; 0 means no limit

	saved string // value when editing started, restored by Cancel
}

func NewInput(value string, limit int) Input {
	return Input{Value: value, Limit: limit}
}

func (in *Input) Begin() { in.saved = in.Value }

func (in *Input) Type(r rune) {
	if in.Limit > 0 && len(in.Value)+len(string(r)) > in.Limit {
		return
	}
	in.Value += string(r)
}

func (in *Input) Backspace() {
	if rs := []rune(in.Value); len(rs) > 0 {
		in.Value = string(rs[:len(rs)-1])
	}
}

func (in *Input) Clear() { in.Value = "" }

func (in *Input) Cancel() { in.Value = in.saved }

// View draws the value; an editing field ends with a caret.
func (in Input) View(editing bool) string {
	if editing {
		return in.Value + "_"
	}
	return in.Value
}

// Select is a fixed option list with one chosen entry. Begin/Step/Cancel give
// it the same edit cycle as Input, so a form can treat both the same way.
type Select struct {
	Options []string
	Index   int

	saved int
}

// NewSelect builds a select around cur. A value that is not in options (an
// older config, a channel the band list does not cover) is appended instead of
// being silently replaced, so loading a form never marks it dirty.
func NewSelect(options []string, cur string) Select {
	s := Select{Options: append([]string(nil), options...)}
	s.Index = s.indexOf(cur)
	return s
}

func (s Select) Value() string {
	if len(s.Options) == 0 {
		return ""
	}
	return s.Options[s.Index]
}

// Set picks v, appending it when the list does not have it yet.
func (s *Select) Set(v string) { s.Index = s.indexOf(v) }

func (s *Select) indexOf(v string) int {
	for i, o := range s.Options {
		if o == v {
			return i
		}
	}
	s.Options = append(s.Options, v)
	return len(s.Options) - 1
}

func (s *Select) Begin() { s.saved = s.Index }

func (s *Select) Step(d int) {
	if len(s.Options) == 0 {
		return
	}
	s.Index = ((s.Index+d)%len(s.Options) + len(s.Options)) % len(s.Options)
}

func (s *Select) Cancel() { s.Index = s.saved }

// View draws the chosen entry; an editing select is bracketed and owns the
// arrow keys.
func (s Select) View(editing bool) string {
	v := s.Value()
	if editing {
		return "[" + v + "]"
	}
	return v
}

// Mask hides a secret of unknown length behind a fixed number of dots.
func Mask(n int) string {
	return strings.Repeat(kit.Glyph.Mask, n)
}

// Buttons is a row of two actions, focusable with the keyboard and clickable in
// the same layout. Focus is the index of the focused action, or -1.
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

// View renders both actions on one line. The focus marker sits in a fixed
// two-cell prefix so the buttons do not shift when the focus moves.
func (b Buttons) View() string {
	var out strings.Builder
	out.WriteString("  ")
	for i := 0; i < 2; i++ {
		if i > 0 {
			out.WriteString("   ")
		}
		label := "[ " + b.label(i) + " ]"
		if i == b.Focus {
			out.WriteString(kit.Glyph.Sel + " " + kit.Accent.Render(label))
			continue
		}
		out.WriteString("  " + kit.Label.Render(label))
	}
	return out.String()
}

// Hits returns the clickable x range of each button relative to the start of
// the row (the same origin View uses). Each range includes 10% of the button
// width as padding on both sides, so a click just outside the label still
// lands on it.
func (b Buttons) Hits() (first, second [2]int) {
	w := func(i int) int { return 2 + lipgloss.Width("[ "+b.Labels[i]+" ]") }
	w0, w1 := w(0), w(1)
	p0, p1 := max(1, w0/10), max(1, w1/10)

	first = [2]int{2 - p0, 2 + w0 - 1 + p0}
	start := 2 + w0 + 3
	second = [2]int{start - p1, start + w1 - 1 + p1}
	return first, second
}
