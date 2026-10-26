package startup

import (
	"fmt"

	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// State is everything the startup pane draws.
type State struct {
	Form    Form
	Path    string
	Active  bool // content has the focus
	Note    string
	Err     string
	Loading bool // the background read has not finished
}

// Pane renders the task list: one switch and one command per task. Comments,
// the shebang and the exit line are kept in the file but not shown.
func Pane(st State) []string {
	head := "  " + style.Label.Render(st.Path)
	if st.Form.Changed() {
		head += "  " + style.Warn.Render("*")
	}
	if st.Loading {
		return []string{"", "  " + style.Dim.Render(style.Tr.Loading)}
	}
	out := []string{"", head}
	if st.Err != "" {
		out = append(out, "  "+style.Bad.Render(st.Err))
	}
	if st.Note != "" {
		out = append(out, "  "+style.Good.Render(st.Note))
	}
	out = append(out, "")

	if !st.Form.Available {
		out = append(out, "  "+style.Bad.Render(fmt.Sprintf(style.Tr.StartupMiss, st.Path)))
		return out
	}

	f := st.Form
	if f.TaskCount() == 0 {
		out = append(out, "  "+style.Dim.Render(style.Tr.StartupNoTasks))
	}
	for row := 0; row < f.TaskCount(); row++ {
		entry, _ := f.TaskAt(row)
		marker := "  "
		if st.Active && row == f.Focus {
			marker = style.Glyph.Sel + " "
		}
		check := components.Checkbox{On: entry.Enabled}.String()

		text := entry.Text
		editing := f.Editing && f.EditIndex() == f.EntryIndex(row)
		if editing {
			text = f.Edit().View(true)
		}

		if st.Active && row == f.Focus {
			out = append(out, "  "+marker+style.Accent.Render(check+" "+text))
			continue
		}
		out = append(out, "  "+marker+check+" "+text)
	}

	buttons := f.Buttons()
	if !st.Active || f.Editing {
		buttons.Focus = -1
	}
	out = append(out, "", buttons.View())
	return out
}

// LineOffset is the body line index of the first task row.
func LineOffset(st State) int {
	off := 3 // blank, header, blank
	if st.Err != "" {
		off++
	}
	if st.Note != "" {
		off++
	}
	return off
}

// RowOffset is the body line index of a focus row: a task index, Apply or
// Cancel. It is what the view has to keep inside the visible window.
func RowOffset(st State) int {
	off := LineOffset(st)
	if st.Form.Focus < st.Form.TaskCount() {
		return off + st.Form.Focus
	}
	return off + max(1, st.Form.TaskCount()) + 1
}

// The switch columns and the button presses are handled by
// components.Checkbox.Click and components.Buttons.Click; the pane only owns
// the row layout.
