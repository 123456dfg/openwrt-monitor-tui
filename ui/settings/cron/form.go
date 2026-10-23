package cron

import (
	"fmt"
	"strconv"
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/widgets/components"
)

// Schedule kinds, in the order the picker lists them. KindCustom keeps an
// expression the four presets cannot describe.
const (
	KindMinutes = iota
	KindHours
	KindDaily
	KindWeekly
	KindCustom
)

// Focus order of the job form.
const (
	FocusKind = iota
	FocusWhen
	FocusWeekday
	FocusCommand
	FocusApply
	FocusCancel
	FocusMax
)

// Form is the new/edited crontab entry. EditLine is the raw crontab line being
// replaced, or -1 when the form describes a new job.
type Form struct {
	Kind    components.Select // options are the localized kind labels
	When    components.Input  // "5", "08:30", or a raw expression for KindCustom
	Weekday components.Select // Sunday-based, the same numbering cron uses
	Command components.Input

	EditLine int

	origSchedule string
	origCommand  string

	Focus   int
	Editing bool
	Busy    bool
}

func NewForm() Form {
	return Form{
		EditLine: -1,
		Kind:     components.NewSelect(kit.Tr.CronKinds[:], kit.Tr.CronKinds[KindMinutes]),
		When:     components.NewInput("", 64),
		Weekday:  components.NewSelect(kit.Tr.CronDays[:], kit.Tr.CronDays[0]),
		Command:  components.NewInput("", 200),
	}
}

// Load fills the form from an existing job so it can be edited in place.
func (f *Form) Load(job collect.CronJob) {
	f.load(job.Line, job.Schedule, job.Command)
	f.Focus = FocusKind
	f.Editing = false
	f.Busy = false
}

func (f *Form) load(line int, schedule, command string) {
	f.EditLine = line
	f.origSchedule = schedule
	f.origCommand = command
	f.Command.Value = command

	kind, when, day, ok := parseSchedule(schedule)
	if !ok {
		kind, when = KindCustom, schedule
	}
	f.Kind.Index = kind
	f.When.Value = when
	f.Weekday.Index = day
}

// Reset drops the edits: back to the loaded job, or an empty form.
func (f *Form) Reset() {
	if f.EditLine >= 0 {
		f.load(f.EditLine, f.origSchedule, f.origCommand)
		f.Editing = false
		return
	}
	fresh := NewForm()
	fresh.Focus = f.Focus
	*f = fresh
}

// Changed reports whether the form would write anything different. Schedules
// are compared by meaning, so reloading "… * * 7" and saving as "… * * 0"
// (both Sunday) is not an edit.
func (f Form) Changed() bool {
	cur, _ := f.Schedule()
	return !sameSchedule(cur, f.origSchedule) || f.CommandLine() != f.origCommand
}

func sameSchedule(a, b string) bool {
	if a == b {
		return true
	}
	ka, wa, da, oka := parseSchedule(a)
	kb, wb, db, okb := parseSchedule(b)
	return oka && okb && ka == kb && wa == wb && da == db
}

// KindIndex identifies the chosen schedule kind.
func (f Form) KindIndex() int { return f.Kind.Index }

// IsNew reports whether Apply would append instead of replace.
func (f Form) IsNew() bool { return f.EditLine < 0 }

// Fields lists the visible form rows in display order. The weekday row only
// makes sense for a weekly schedule, so it disappears for the interval kinds.
func (f Form) Fields() []int {
	switch f.KindIndex() {
	case KindMinutes, KindHours:
		return []int{FocusKind, FocusWhen, FocusCommand}
	}
	return []int{FocusKind, FocusWhen, FocusWeekday, FocusCommand}
}

// FieldRow is the display index of a focus constant, or -1 when the row is
// hidden for the current kind.
func (f Form) FieldRow(focus int) int {
	for i, r := range f.Fields() {
		if r == focus {
			return i
		}
	}
	return -1
}

// StepFocus moves the cursor through the visible rows, then the buttons.
func (f *Form) StepFocus(d int) {
	rows := append(f.Fields(), FocusApply, FocusCancel)
	cur := 0
	for i, r := range rows {
		if r == f.Focus {
			cur = i
			break
		}
	}
	f.Focus = rows[((cur+d)%len(rows)+len(rows))%len(rows)]
}

// NormalizeFocus moves the cursor off a row that just got hidden.
func (f *Form) NormalizeFocus() {
	if f.FieldRow(f.Focus) >= 0 || f.Focus == FocusApply || f.Focus == FocusCancel {
		return
	}
	f.Focus = FocusCommand
}

// WhenLabel is the label of the value row.
func (f Form) WhenLabel() string {
	switch f.KindIndex() {
	case KindDaily, KindWeekly:
		return kit.Tr.CronAt
	case KindCustom:
		return kit.Tr.CronExpr
	}
	return kit.Tr.CronEvery
}

// WhenHint documents the expected input for the chosen kind.
func (f Form) WhenHint() string {
	switch f.KindIndex() {
	case KindMinutes:
		return "1-59"
	case KindHours:
		return "1-23"
	case KindCustom:
		return "5 fields"
	}
	return "HH:MM"
}

// Schedule turns the inputs into cron's five fields.
func (f Form) Schedule() (string, string) {
	switch f.KindIndex() {
	case KindMinutes:
		n, ok := number(f.When.Value, 1, 59)
		if !ok {
			return "", kit.Tr.ValCronRange
		}
		return fmt.Sprintf("*/%d * * * *", n), ""
	case KindHours:
		n, ok := number(f.When.Value, 1, 23)
		if !ok {
			return "", kit.Tr.ValCronRange
		}
		return fmt.Sprintf("0 */%d * * *", n), ""
	case KindDaily:
		h, m, ok := clock(f.When.Value)
		if !ok {
			return "", kit.Tr.ValCronTime
		}
		return fmt.Sprintf("%d %d * * *", m, h), ""
	case KindWeekly:
		h, m, ok := clock(f.When.Value)
		if !ok {
			return "", kit.Tr.ValCronTime
		}
		return fmt.Sprintf("%d %d * * %d", m, h, f.Weekday.Index), ""
	}
	expr := strings.TrimSpace(f.When.Value)
	if len(strings.Fields(expr)) != 5 {
		return "", kit.Tr.ValCronExpr
	}
	return expr, ""
}

func (f Form) Validate() string {
	if _, msg := f.Schedule(); msg != "" {
		return msg
	}
	cmd := strings.TrimSpace(f.Command.Value)
	if cmd == "" || strings.ContainsAny(cmd, "\r\n") {
		return kit.Tr.ValCommand
	}
	return ""
}

// CommitEdit closes the row, trimming the input the way crontab lines are
// compared and written.
func (f *Form) CommitEdit() {
	switch f.Focus {
	case FocusWhen:
		f.When.Value = strings.TrimSpace(f.When.Value)
	case FocusCommand:
		f.Command.Value = strings.TrimSpace(f.Command.Value)
	}
	f.Editing = false
}

// CommandLine is the command as it will be written.
func (f Form) CommandLine() string { return strings.TrimSpace(f.Command.Value) }

// Origin is the exact job the form was loaded from, used to detect a file that
// changed underneath the editor.
func (f Form) Origin() string {
	if f.EditLine < 0 {
		return ""
	}
	return f.origSchedule + " " + f.origCommand
}

// Text is the input widget of the focused row, or nil.
func (f *Form) Text() *components.Input {
	switch f.Focus {
	case FocusWhen:
		return &f.When
	case FocusCommand:
		return &f.Command
	}
	return nil
}

// Picker is the select widget of the focused row, or nil.
func (f *Form) Picker() *components.Select {
	switch f.Focus {
	case FocusKind:
		return &f.Kind
	case FocusWeekday:
		return &f.Weekday
	}
	return nil
}

// Buttons describes the action row for the shared button widget.
func (f Form) Buttons() components.Buttons {
	return components.Buttons{
		Labels:    [2]string{kit.Tr.Apply, kit.Tr.Cancel},
		Focus:     buttonIndex(f.Focus),
		Busy:      f.Busy,
		BusyLabel: kit.Tr.Applying,
	}
}

func buttonIndex(focus int) int {
	switch focus {
	case FocusApply:
		return 0
	case FocusCancel:
		return 1
	}
	return -1
}

// parseSchedule maps a five-field expression back to one of the presets. It
// only accepts the shapes Schedule produces; anything else stays custom.
func parseSchedule(expr string) (kind int, when string, day int, ok bool) {
	f := strings.Fields(expr)
	if len(f) != 5 || f[2] != "*" || f[3] != "*" {
		return 0, "", 0, false
	}
	if n, isEvery := everyN(f[0]); isEvery && f[1] == "*" && f[4] == "*" {
		if n < 1 || n > 59 {
			return 0, "", 0, false
		}
		return KindMinutes, strconv.Itoa(n), 0, true
	}
	if f[0] == "0" && f[4] == "*" {
		if n, isEvery := everyN(f[1]); isEvery {
			if n < 1 || n > 23 {
				return 0, "", 0, false
			}
			return KindHours, strconv.Itoa(n), 0, true
		}
	}
	m, ok1 := number(f[0], 0, 59)
	h, ok2 := number(f[1], 0, 23)
	if !ok1 || !ok2 {
		return 0, "", 0, false
	}
	when = fmt.Sprintf("%02d:%02d", h, m)
	if f[4] == "*" {
		return KindDaily, when, 0, true
	}
	d, ok := number(f[4], 0, 7)
	if !ok {
		return 0, "", 0, false
	}
	if d == 7 {
		d = 0
	}
	return KindWeekly, when, d, true
}

// everyN accepts "*/n" and returns n.
func everyN(s string) (int, bool) {
	rest, found := strings.CutPrefix(s, "*/")
	if !found {
		return 0, false
	}
	n, err := strconv.Atoi(rest)
	if err != nil {
		return 0, false
	}
	return n, true
}

func number(s string, lo, hi int) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < lo || n > hi {
		return 0, false
	}
	return n, true
}

func clock(s string) (h, m int, ok bool) {
	hh, mm, found := strings.Cut(strings.TrimSpace(s), ":")
	if !found {
		return 0, 0, false
	}
	h, ok1 := number(hh, 0, 23)
	m, ok2 := number(mm, 0, 59)
	return h, m, ok1 && ok2
}
