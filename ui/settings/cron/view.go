package cron

import (
	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
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
		out = append(out, "  "+style.Bad.Render(st.Err))
	}
	if st.Note != "" {
		out = append(out, "  "+style.Good.Render(st.Note))
	}
	if st.Err != "" || st.Note != "" {
		out = append(out, "")
	}
	if st.Tab == TabList {
		if st.Loading {
			return append(out, "  "+style.Dim.Render(style.Tr.Loading))
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
	names := tabs(st)
	return components.Tabs{Names: names[:], Cur: cur}.View()
}

func tabs(st State) [2]string {
	editor := style.Tr.CronTabNew
	if st.Form.EditLine >= 0 {
		editor = style.Tr.CronTabEdit
	}
	return [2]string{style.Tr.CronTabList, editor}
}

// TabClick maps a press on the strip (x in pane-content coordinates) to a
// sub-tab index; -1 when nothing was hit.
func TabClick(st State, x int) int {
	cur := st.Tab
	if cur < TabList || cur > TabEditor {
		cur = TabList
	}
	names := tabs(st)
	return components.Tabs{Names: names[:], Cur: cur}.Click(style.SidebarW+4, x)
}

// TabHits returns the clickable x range of both sub-tabs (layout helper for
// tests; clicks go through TabClick).
func TabHits(st State) [][2]int {
	cur := st.Tab
	if cur < TabList || cur > TabEditor {
		cur = TabList
	}
	names := tabs(st)
	// Sidebar width + pane indent + the strip's own two-space prefix.
	return components.Tabs{Names: names[:], Cur: cur}.Hits(style.SidebarW + 4)
}

func list(st State) []string {
	switch {
	case !st.Avail:
		return []string{"  " + style.Dim.Render(style.Tr.CronNoCron)}
	case len(st.Jobs) == 0:
		return []string{"  " + style.Dim.Render(style.Tr.CronNone)}
	}
	out := make([]string, 0, len(st.Jobs))
	for i, job := range st.Jobs {
		// Only the command and its switch are shown; the schedule lives
		// behind Enter.
		check := components.Checkbox{On: job.Enabled}.String()
		if st.Active && i == st.Cursor {
			out = append(out, "  "+style.Glyph.Sel+" "+style.Accent.Render(check+" "+job.Command))
			continue
		}
		out = append(out, "  "+"  "+check+" "+job.Command)
	}
	return out
}

// The job switch is a components.Checkbox drawn at columns 4..6 of the row;
// the component resolves the press, the pane persists the flip.

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
			name, value = style.Tr.CronKind, f.Kind.View(editing(FocusKind))
		case FocusWhen:
			name, value = f.WhenLabel(), f.When.View(editing(FocusWhen))
		case FocusWeekday:
			name, value = style.Tr.CronWeekday, f.Weekday.View(editing(FocusWeekday))
			dim = f.KindIndex() != KindWeekly
		case FocusCommand:
			name, value = style.Tr.CronCommand, f.Command.View(editing(FocusCommand))
		}
		line := fieldRow(row == focus, name, value, dim)
		if row == focus {
			switch row {
			case FocusWhen:
				line += "   " + style.Dim.Render(f.WhenHint())
			case FocusKind, FocusWeekday:
				if f.Editing {
					line += "   " + style.Dim.Render("←/→")
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
	nameStyle, valStyle := style.Label, style.Style()
	if dim {
		nameStyle, valStyle = style.Dim, style.Dim
	}
	if focused {
		marker = style.Glyph.Sel + " "
		if !dim {
			nameStyle, valStyle = style.Accent, style.Accent
		}
	}
	return "  " + marker + nameStyle.Render(style.PadRight(name, 9)) + " " + valStyle.Render(value)
}
