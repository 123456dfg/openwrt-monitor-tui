package components

import (
	"strings"
	"testing"
)

func TestCheckboxClick(t *testing.T) {
	c := Checkbox{On: false}
	if !c.Click(0, 0) || !c.On {
		t.Fatalf("click did not toggle: %+v", c)
	}
	c.Click(3, 1) // coordinates are ignored, any press counts
	if c.On {
		t.Fatal("second click did not toggle back")
	}
}

func TestRadioClickAndWheel(t *testing.T) {
	r := Radio{Options: []string{"a", "b", "c"}, Cursor: 0, Sel: 0}
	if !r.Click(0, 2) || r.Cursor != 2 || r.Sel != 2 {
		t.Fatalf("click row 2: cursor=%d sel=%d", r.Cursor, r.Sel)
	}
	if r.Click(0, 9) {
		t.Fatal("out-of-range row consumed")
	}
	r.Wheel(-1) // previews without selecting
	if r.Cursor != 1 || r.Sel != 2 {
		t.Fatalf("wheel: cursor=%d sel=%d", r.Cursor, r.Sel)
	}
}

func TestInputClickBegins(t *testing.T) {
	in := NewInput("abc", 0)
	if !in.Click(0, 0) {
		t.Fatal("click not consumed")
	}
	in.Type('x')
	in.Cancel() // restores the value from Begin, proving Begin ran
	if in.Value != "abc" {
		t.Fatalf("click did not begin an editing session: %q", in.Value)
	}
}

func TestSelectWheel(t *testing.T) {
	s := NewSelect([]string{"auto", "1", "6"}, "auto")
	s.Wheel(+1)
	if s.Value() != "1" {
		t.Fatalf("wheel: %q", s.Value())
	}
	s.Wheel(-1)
	if s.Value() != "auto" {
		t.Fatalf("wheel back: %q", s.Value())
	}
}

func TestButtonsClick(t *testing.T) {
	b := Buttons{Labels: [2]string{"Apply", "Cancel"}}
	if b.Click(0, 0) {
		t.Fatal("press on the leading indent consumed")
	}
	if !b.Click(4, 0) || b.Focus != 0 {
		t.Fatalf("apply click: focus=%d", b.Focus)
	}
	hits := b.Hits()
	if !b.Click(hits[1][0], 0) || b.Focus != 1 {
		t.Fatalf("cancel click: focus=%d", b.Focus)
	}
}

func TestSingleButton(t *testing.T) {
	b := Buttons{Labels: [2]string{"Start", ""}}
	if len(b.Hits()) != 1 {
		t.Fatalf("single button rendered %d hit ranges", len(b.Hits()))
	}
	if strings.Contains(b.View(), "[  ]") {
		t.Fatalf("hidden button rendered: %q", b.View())
	}
	if !b.Click(4, 0) || b.Focus != 0 {
		t.Fatalf("single click: focus=%d", b.Focus)
	}
}

func TestTabsClickAndWheel(t *testing.T) {
	tabs := Tabs{Names: []string{"AB", "CDEF"}, Cur: 0}
	if i := tabs.Click(2, 2); i != 0 {
		t.Fatalf("first tab: %d", i)
	}
	if i := tabs.Click(2, 7); i != 1 {
		t.Fatalf("second tab: %d", i)
	}
	if i := tabs.Click(2, 60); i != -1 {
		t.Fatalf("miss: %d", i)
	}
	if !tabs.Wheel(+1) || tabs.Cur != 1 {
		t.Fatalf("wheel: %d", tabs.Cur)
	}
}
