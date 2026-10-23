package components

import (
	"strings"

	"openwrt-monitor-tui/ui/kit"
)

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

// Click starts an editing session on the field.
func (in *Input) Click(x, y int) bool {
	in.Begin()
	return true
}

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

// Mask hides a secret of unknown length behind a fixed number of dots.
func Mask(n int) string {
	return strings.Repeat(kit.Glyph.Mask, n)
}
