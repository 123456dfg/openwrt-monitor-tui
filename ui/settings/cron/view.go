package cron

import (
	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/widgets"
)

// Sub-tabs of the cron pane.
const (
	TabList = iota
	TabEditor
)

// State is everything the cron pane draws.
type State struct {
	Form    Form
	Tab     int
	Cursor  int // selected job on the list tab
	Jobs    []collect.CronJob
	Avail   bool // a crontab (or /etc/crontabs) exists
	Loading bool // the background read has not finished
	Active  bool // content has the focus
	Note    string
	Err     string
}

// Pane renders the tab strip, then either the job list or the form.
func Pane(st State) []string {
	out := []string{"  " + strip(st), ""}
	if st.Err != "" {
		out = append(out, "  "+kit.Bad.Render(st.Err))
	}
	if st.Note != "" {
		out = append(out, "  "+kit.Good.Render(st.Note))
	}
	if st.Err != "" || st.Note != "" {
		out = append(out, "")
	}
	if st.Tab == TabList {
		if st.Loading {
			return append(out, "  "+kit.Dim.Render(kit.Tr.Loading))
		}
		return append(out, list(st)...)
	}
	return append(out, rows(st)...)
}

// ContentOffset is the body line index where the list or form starts.
func ContentOffset(st State) int {
	off := 2 // tab strip + blank line
	if st.Err != "" {
		off++
	}
	if st.Note != "" {
		off++
	}
	if st.Err != "" || st.Note != "" {
		off++
	}
	return off
}

// FormRowOffset is the body line index of the first form row.
func FormRowOffset(st State) int { return ContentOffset(st) }

// ButtonsRow is the body line index of the action row.
func ButtonsRow(st State) int {
	return ContentOffset(st) + len(st.Form.Fields()) + 1
}

func strip(st State) string {
	cur := st.Tab
	if cur < TabList || cur > TabEditor {
		cur = TabList
	}
	line := ""
	for i, name := range tabs(st) {
		if i > 0 {
			line += "   "
		}
		if i == cur {
			if kit.ColorOn() {
				line += kit.Accent.Render(name)
			} else {
				line += "[" + name + "]"
			}
			continue
		}
		line += kit.Label.Render(name)
	}
	return line
}

func tabs(st State) [2]string {
	editor := kit.Tr.CronTabNew
	if st.Form.EditLine >= 0 {
		editor = kit.Tr.CronTabEdit
	}
	return [2]string{kit.Tr.CronTabList, editor}
}

// TabHits returns the clickable x range of both sub-tabs.
func TabHits(st State) [2][2]int {
	var hits [2][2]int
	x := kit.SidebarW + 4 // sidebar + pane indent
	for i, name := range tabs(st) {
		w := lipgloss.Width(name)
		hits[i] = [2]int{x, x + w - 1}
		x += w + 3
	}
	return hits
}

func list(st State) []string {
	switch {
	case !st.Avail:
		return []string{"  " + kit.Dim.Render(kit.Tr.CronNoCron)}
	case len(st.Jobs) == 0:
		return []string{"  " + kit.Dim.Render(kit.Tr.CronNone)}
	}
	out := make([]string, 0, len(st.Jobs))
	for i, job := range st.Jobs {
		// Only the command and its switch are shown; the schedule lives
		// behind Enter.
		check := widgets.Checkbox{On: job.Enabled}.String()
		if st.Active && i == st.Cursor {
			out = append(out, "  "+kit.Glyph.Sel+" "+kit.Accent.Render(check+" "+job.Command))
			continue
		}
		out = append(out, "  "+"  "+check+" "+job.Command)
	}
	return out
}

// CheckHits is the x range of the switch inside a job row, relative to the
// start of the pane content.
func CheckHits() [2]int { return [2]int{4, 6} }

func rows(st State) []string {
	f := st.Form
	focus := f.Focus
	if !st.Active {
		focus = -1
	}
	editing := func(i int) bool { return f.Editing && i == focus }

	visible := f.Fields()
	out := make([]string, 0, len(visible)+2)
	for _, row := range visible {
		name, value, dim := "", "", false
		switch row {
		case FocusKind:
			name, value = kit.Tr.CronKind, f.Kind.View(editing(FocusKind))
		case FocusWhen:
			name, value = f.WhenLabel(), f.When.View(editing(FocusWhen))
		case FocusWeekday:
			name, value = kit.Tr.CronWeekday, f.Weekday.View(editing(FocusWeekday))
			dim = f.KindIndex() != KindWeekly
		case FocusCommand:
			name, value = kit.Tr.CronCommand, f.Command.View(editing(FocusCommand))
		}
		line := fieldRow(row == focus, name, value, dim)
		if row == focus {
			switch row {
			case FocusWhen:
				line += "   " + kit.Dim.Render(f.WhenHint())
			case FocusKind, FocusWeekday:
				if f.Editing {
					line += "   " + kit.Dim.Render("←/→")
				}
			}
		}
		out = append(out, line)
	}
	buttons := f.Buttons()
	if !st.Active || f.Editing {
		buttons.Focus = -1
	}
	out = append(out, "", buttons.View())
	return out
}

func fieldRow(focused bool, name, value string, dim bool) string {
	marker := "  "
	nameStyle, valStyle := kit.Label, kit.Style()
	if dim {
		nameStyle, valStyle = kit.Dim, kit.Dim
	}
	if focused {
		marker = kit.Glyph.Sel + " "
		if !dim {
			nameStyle, valStyle = kit.Accent, kit.Accent
		}
	}
	return "  " + marker + nameStyle.Render(kit.PadRight(name, 9)) + " " + valStyle.Render(value)
}
