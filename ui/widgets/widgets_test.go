package widgets

import (
	"strings"
	"testing"

	"openwrt-monitor-tui/ui/kit"
)

func TestCheckbox(t *testing.T) {
	c := Checkbox{On: true}
	if c.String() != "[x]" {
		t.Fatalf("on: %q", c.String())
	}
	c.Toggle()
	if c.On || c.String() != "[ ]" {
		t.Fatalf("toggled: %v %q", c.On, c.String())
	}
}

func TestInput(t *testing.T) {
	in := NewInput("abc", 4)
	in.Begin()
	in.Type('d')
	in.Type('e') // over the limit
	if in.Value != "abcd" {
		t.Fatalf("limit: %q", in.Value)
	}
	if got := in.View(true); got != "abcd_" {
		t.Fatalf("editing view: %q", got)
	}
	in.Backspace()
	if in.Value != "abc" {
		t.Fatalf("backspace: %q", in.Value)
	}
	in.Cancel()
	if in.Value != "abc" || in.View(false) != "abc" {
		t.Fatalf("cancel: %q", in.Value)
	}
	in.Clear()
	if in.Value != "" {
		t.Fatalf("clear: %q", in.Value)
	}
}

func TestSelect(t *testing.T) {
	// a value outside the list is kept, so loading never marks the form dirty
	s := NewSelect([]string{"auto", "1", "6"}, "14")
	if s.Value() != "14" {
		t.Fatalf("unknown value dropped: %q", s.Value())
	}
	s.Step(+1)
	if s.Value() != "auto" {
		t.Fatalf("wrap: %q", s.Value())
	}
	s.Begin()
	s.Step(+1)
	if s.Value() != "1" {
		t.Fatalf("step: %q", s.Value())
	}
	s.Cancel()
	if s.Value() != "auto" {
		t.Fatalf("cancel: %q", s.Value())
	}
	if got := s.View(true); got != "[auto]" {
		t.Fatalf("editing view: %q", got)
	}
	s.Set("6")
	if s.Value() != "6" || s.View(false) != "6" {
		t.Fatalf("set: %q", s.Value())
	}

	empty := NewSelect(nil, "")
	empty.Step(+1) // must not panic or wrap into an invalid index
	if empty.Value() != "" {
		t.Fatalf("empty: %q", empty.Value())
	}
}

func TestRadio(t *testing.T) {
	kit.SetColor(false)
	t.Cleanup(func() { kit.SetColor(true) })

	r := NewRadio(1)
	r.Options = []string{"auto", "en", "zh"}

	lines := r.Lines([]string{"Auto", "English", "中文"})
	if len(lines) != 3 {
		t.Fatalf("lines: %d", len(lines))
	}
	if !strings.Contains(lines[1], "(o)") || !strings.Contains(lines[1], kit.Glyph.Sel) {
		t.Fatalf("selected line: %q", lines[1])
	}
	for i, line := range lines {
		if i != 1 && strings.Contains(line, "(o)") {
			t.Fatalf("line %d marked: %q", i, line)
		}
	}

	r.Move(+1)
	if r.Cursor != 2 || r.Sel != 1 {
		t.Fatalf("move previews without selecting: cursor=%d sel=%d", r.Cursor, r.Sel)
	}
	r.Choose()
	if r.Sel != 2 {
		t.Fatalf("choose: %d", r.Sel)
	}
	r.Move(+1) // wraps
	if r.Cursor != 0 {
		t.Fatalf("wrap: %d", r.Cursor)
	}
	r.Cursor = -1
	if strings.Contains(strings.Join(r.Lines([]string{"a", "b", "c"}), "\n"), kit.Glyph.Sel) {
		t.Fatal("negative cursor should hide the marker")
	}
}

func TestMask(t *testing.T) {
	kit.SetColor(false)
	t.Cleanup(func() { kit.SetColor(true) })
	if got := Mask(4); strings.Count(got, kit.Glyph.Mask) != 4 {
		t.Fatalf("mask: %q", got)
	}
}
