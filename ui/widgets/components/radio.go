package components

import "openwrt-monitor-tui/ui/style"

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

// Click selects the option on row y (the row index within the rendered list).
func (r *Radio) Click(x, y int) bool {
	if y < 0 || y >= len(r.Options) {
		return false
	}
	r.Cursor, r.Sel = y, y
	return true
}

// Wheel previews the previous/next option without selecting it.
func (r *Radio) Wheel(d int) bool {
	r.Move(d)
	return true
}

// Lines renders one line per option: a cursor marker, a (o)/( ) box and the
// label. A negative Cursor hides the marker, which is what a previewed but not
// entered pane wants.
func (r Radio) Lines(labels []string) []string {
	out := make([]string, 0, len(labels))
	for i, label := range labels {
		marker, row := "  ", style.Label
		if i == r.Cursor {
			marker, row = style.Glyph.Sel+" ", style.Accent
		}
		box := "( )"
		if i == r.Sel {
			box = "(o)"
		}
		out = append(out, marker+row.Render(box+" "+label))
	}
	return out
}
