package startup

import (
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/widgets"
)

// lineLimit caps one command; rc.local is a boot script, not a buffer.
const lineLimit = 400

// offPrefix marks a task that is kept in the file but disabled. It is a comment
// as far as the shell is concerned.
const offPrefix = collect.OffPrefix

// Entry is one line of /etc/rc.local. Only Task entries are shown: comments,
// the shebang, blank lines and the final exit are kept verbatim but hidden.
type Entry struct {
	Text    string
	Task    bool
	Enabled bool
}

// Form is an editable copy of /etc/rc.local. Org is what the file held when it
// was read: it drives the dirty marker, Cancel and the refresh from the sampler.
type Form struct {
	Available bool
	Entries   []Entry
	Org       []Entry

	Focus   int // task index, then Apply, then Cancel
	Editing bool
	Busy    bool

	edit    widgets.Input
	editIdx int // entry index being edited
}

func FromInfo(info collect.StartupInfo, focus int) Form {
	f := Form{Available: info.Available}
	f.Entries = parseLines(info.Content)
	f.Org = clone(f.Entries)
	f.Focus = focus
	f.ClampFocus()
	return f
}

func clone(in []Entry) []Entry {
	return append([]Entry(nil), in...)
}

// parseLines keeps every line but marks the ones the pane shows as tasks:
// commands that are neither comments, blank nor the script's exit.
func parseLines(lines []string) []Entry {
	out := make([]Entry, 0, len(lines))
	for _, raw := range lines {
		text := strings.TrimSpace(raw)
		enabled := true
		if rest, off := strings.CutPrefix(text, offPrefix); off {
			text, enabled = strings.TrimSpace(rest), false
		}
		if text == "" || strings.HasPrefix(text, "#") || isExit(text) {
			out = append(out, Entry{Text: raw})
			continue
		}
		out = append(out, Entry{Text: text, Task: true, Enabled: enabled})
	}
	return out
}

// isExit reports whether the line ends the script. It stays hidden and is
// always written back, so appended tasks stay above it.
func isExit(text string) bool {
	return text == "exit" || strings.HasPrefix(text, "exit ")
}

func (f Form) Changed() bool {
	if len(f.Entries) != len(f.Org) {
		return true
	}
	for i := range f.Entries {
		if f.Entries[i] != f.Org[i] {
			return true
		}
	}
	return false
}

// Tasks returns the entry indices that are shown as tasks, in file order.
func (f Form) Tasks() []int {
	idx := make([]int, 0, len(f.Entries))
	for i, e := range f.Entries {
		if e.Task {
			idx = append(idx, i)
		}
	}
	return idx
}

func (f Form) TaskCount() int { return len(f.Tasks()) }

// ApplyIndex and CancelIndex are the focus values of the two buttons.
func (f Form) ApplyIndex() int  { return f.TaskCount() }
func (f Form) CancelIndex() int { return f.TaskCount() + 1 }
func (f Form) FocusCount() int  { return f.TaskCount() + 2 }

// EntryIndex maps a task row to its entry.
func (f Form) EntryIndex(task int) int {
	idx := f.Tasks()
	if task < 0 || task >= len(idx) {
		return -1
	}
	return idx[task]
}

// TaskAt returns the entry shown at a task row.
func (f Form) TaskAt(task int) (Entry, bool) {
	i := f.EntryIndex(task)
	if i < 0 {
		return Entry{}, false
	}
	return f.Entries[i], true
}

func (f *Form) ClampFocus() {
	f.Focus = min(max(f.Focus, 0), f.FocusCount()-1)
}

// Reset drops the edits and goes back to the file as read.
func (f *Form) Reset() {
	f.Entries = clone(f.Org)
	f.Editing = false
	f.ClampFocus()
}

// Adopt makes the current entries the new baseline after a successful save.
func (f *Form) Adopt() { f.Org = clone(f.Entries) }

// Toggle flips the switch of the task at the given row.
func (f *Form) Toggle(task int) bool {
	i := f.EntryIndex(task)
	if i < 0 {
		return false
	}
	f.Entries[i].Enabled = !f.Entries[i].Enabled
	return true
}

// Begin edits the task at the given row.
func (f *Form) Begin(task int) bool {
	i := f.EntryIndex(task)
	if i < 0 {
		return false
	}
	f.edit = widgets.NewInput(f.Entries[i].Text, lineLimit)
	f.edit.Begin()
	f.editIdx = i
	f.Editing = true
	return true
}

// CommitEdit writes the edited text back and leaves the row.
func (f *Form) CommitEdit() {
	if f.Editing && f.editIdx >= 0 && f.editIdx < len(f.Entries) {
		f.Entries[f.editIdx].Text = strings.TrimRight(f.edit.Value, " ")
	}
	f.Editing = false
}

// Insert adds a task above the trailing exit (or at the end) and edits it.
func (f *Form) Insert() {
	at := len(f.Entries)
	for i := len(f.Entries) - 1; i >= 0; i-- {
		if strings.TrimSpace(f.Entries[i].Text) == "" {
			continue
		}
		if !f.Entries[i].Task && isExit(strings.TrimSpace(f.Entries[i].Text)) {
			at = i
		}
		break
	}
	entry := Entry{Task: true, Enabled: true}
	f.Entries = append(f.Entries, Entry{})
	copy(f.Entries[at+1:], f.Entries[at:])
	f.Entries[at] = entry

	task := 0
	for i := 0; i < at; i++ {
		if f.Entries[i].Task {
			task++
		}
	}
	f.Focus = task
	f.Begin(task)
	return
}

// Delete removes the task at the given row.
func (f *Form) Delete(task int) {
	i := f.EntryIndex(task)
	if i < 0 {
		return
	}
	f.Entries = append(f.Entries[:i], f.Entries[i+1:]...)
	if f.Focus >= f.TaskCount() {
		f.Focus = f.TaskCount() - 1
	}
	if f.Focus < 0 {
		f.Focus = 0
	}
}

// Edit is the input owned by the task being edited.
func (f *Form) Edit() *widgets.Input { return &f.edit }

// EditIndex is the entry currently being edited.
func (f Form) EditIndex() int { return f.editIdx }

// Body is the file text as it would be written.
func (f Form) Body() string {
	if len(f.Entries) == 0 {
		return ""
	}
	lines := make([]string, len(f.Entries))
	for i, e := range f.Entries {
		line := e.Text
		if e.Task && !e.Enabled {
			line = offPrefix + line
		}
		lines[i] = line
	}
	return strings.Join(lines, "\n") + "\n"
}

// Buttons describes the action row for the shared button widget.
func (f Form) Buttons() widgets.Buttons {
	return widgets.Buttons{
		Labels:    [2]string{kit.Tr.Apply, kit.Tr.Cancel},
		Focus:     buttonIndex(f),
		Busy:      f.Busy,
		BusyLabel: kit.Tr.Applying,
	}
}

func buttonIndex(f Form) int {
	switch f.Focus {
	case f.ApplyIndex():
		return 0
	case f.CancelIndex():
		return 1
	}
	return -1
}
