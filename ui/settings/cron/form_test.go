package cron

import (
	"os"
	"path/filepath"
	"testing"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
)

func TestScheduleKinds(t *testing.T) {
	kit.SetLang("en")
	f := NewForm()

	f.Kind.Index = KindMinutes
	f.When.Value = "5"
	if sched, msg := f.Schedule(); sched != "*/5 * * * *" || msg != "" {
		t.Fatalf("minutes: %q %q", sched, msg)
	}

	f.Kind.Index = KindHours
	f.When.Value = "2"
	if sched, _ := f.Schedule(); sched != "0 */2 * * *" {
		t.Fatalf("hours: %q", sched)
	}

	f.Kind.Index = KindDaily
	f.When.Value = "08:30"
	if sched, _ := f.Schedule(); sched != "30 8 * * *" {
		t.Fatalf("daily: %q", sched)
	}

	f.Kind.Index = KindWeekly
	f.When.Value = "23:59"
	f.Weekday.Index = 1
	if sched, _ := f.Schedule(); sched != "59 23 * * 1" {
		t.Fatalf("weekly: %q", sched)
	}

	// the label follows the kind
	f.Kind.Index = KindDaily
	if f.WhenLabel() != kit.Tr.CronAt {
		t.Fatalf("label: %q", f.WhenLabel())
	}
	f.Kind.Index = KindMinutes
	if f.WhenLabel() != kit.Tr.CronEvery {
		t.Fatalf("label: %q", f.WhenLabel())
	}
}

func TestValidate(t *testing.T) {
	kit.SetLang("en")
	f := NewForm()
	f.When.Value = "0"
	f.Command.Value = "/bin/true"
	if f.Validate() != kit.Tr.ValCronRange {
		t.Fatal("zero interval accepted")
	}
	f.When.Value = "60"
	if f.Validate() != kit.Tr.ValCronRange {
		t.Fatal("60 minutes accepted")
	}

	f.Kind.Index = KindDaily
	f.When.Value = "25:00"
	if f.Validate() != kit.Tr.ValCronTime {
		t.Fatal("invalid hour accepted")
	}
	f.When.Value = "8-30"
	if f.Validate() != kit.Tr.ValCronTime {
		t.Fatal("missing colon accepted")
	}
	f.When.Value = "08:30"
	f.Command.Value = "   "
	if f.Validate() != kit.Tr.ValCommand {
		t.Fatal("empty command accepted")
	}
	f.Command.Value = " /bin/true "
	if msg := f.Validate(); msg != "" {
		t.Fatalf("valid job rejected: %s", msg)
	}
	if f.CommandLine() != "/bin/true" {
		t.Fatalf("command line: %q", f.CommandLine())
	}

	// hours are capped at 23, so 24 must fail
	f.Kind.Index = KindHours
	f.When.Value = "24"
	if f.Validate() != kit.Tr.ValCronRange {
		t.Fatal("24 hours accepted")
	}
}

func TestReset(t *testing.T) {
	f := NewForm()
	f.Kind.Index = KindDaily
	f.When.Value = "08:30"
	f.Command.Value = "/bin/true"
	f.Focus = FocusCommand
	f.Editing = true
	f.Reset()
	if f.When.Value != "" || f.Command.Value != "" || f.Editing || f.Focus != FocusCommand {
		t.Fatalf("reset: %+v", f)
	}
	if f.Kind.Index != KindMinutes {
		t.Fatalf("kind after reset: %d", f.Kind.Index)
	}
}

func TestAddWritesLine(t *testing.T) {
	kit.SetLang("en")
	path := filepath.Join(t.TempDir(), "root")
	if err := os.WriteFile(path, []byte("# existing\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldPath, oldReload := cronPath, cronReload
	cronPath = func() string { return path }
	reloaded := 0
	cronReload = func() error { reloaded++; return nil }
	t.Cleanup(func() { cronPath, cronReload = oldPath, oldReload })

	if err := Add("*/5 * * * *", "/usr/bin/logger hi"); err != nil {
		t.Fatalf("add: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "# existing\n*/5 * * * * /usr/bin/logger hi\n"
	if string(raw) != want {
		t.Fatalf("file: %q", raw)
	}
	if reloaded != 1 {
		t.Fatalf("reload calls: %d", reloaded)
	}
}

func TestParseScheduleRoundTrip(t *testing.T) {
	kit.SetLang("en")
	cases := []struct {
		expr string
		kind int
		when string
		day  int
	}{
		{"*/5 * * * *", KindMinutes, "5", 0},
		{"0 */2 * * *", KindHours, "2", 0},
		{"30 8 * * *", KindDaily, "08:30", 0},
		{"59 23 * * 6", KindWeekly, "23:59", 6},
	}
	for _, c := range cases {
		f := NewForm()
		f.Load(collect.CronJob{Schedule: c.expr, Command: "/bin/true", Line: 4})
		if f.KindIndex() != c.kind || f.When.Value != c.when || f.Weekday.Index != c.day {
			t.Fatalf("%q -> kind=%d when=%q day=%d", c.expr, f.KindIndex(), f.When.Value, f.Weekday.Index)
		}
		if f.EditLine != 4 || f.Changed() {
			t.Fatalf("%q: line=%d changed=%v", c.expr, f.EditLine, f.Changed())
		}
		if sched, msg := f.Schedule(); sched != c.expr || msg != "" {
			t.Fatalf("%q round trip: %q %q", c.expr, sched, msg)
		}
	}

	// 7 is Sunday as well; loading it must not look like an edit
	sun := NewForm()
	sun.Load(collect.CronJob{Schedule: "59 23 * * 7", Command: "/bin/true", Line: 1})
	if sun.Weekday.Index != 0 || sun.Changed() {
		t.Fatalf("day 7: index=%d changed=%v", sun.Weekday.Index, sun.Changed())
	}
}

func TestLoadCustomSchedule(t *testing.T) {
	kit.SetLang("en")
	f := NewForm()
	f.Load(collect.CronJob{Schedule: "*/5 3 * * *", Command: "x", Line: 0})
	if f.KindIndex() != KindCustom || f.When.Value != "*/5 3 * * *" {
		t.Fatalf("custom: kind=%d when=%q", f.KindIndex(), f.When.Value)
	}
	if sched, _ := f.Schedule(); sched != "*/5 3 * * *" {
		t.Fatalf("custom schedule: %q", sched)
	}
	if f.Changed() {
		t.Fatal("custom job loaded as changed")
	}
	// an incomplete expression is rejected once it is edited
	f.When.Value = "*/5 3"
	if msg := f.Validate(); msg != kit.Tr.ValCronExpr {
		t.Fatalf("validation: %q", msg)
	}
}

func TestLoadedFormReset(t *testing.T) {
	kit.SetLang("en")
	f := NewForm()
	f.Load(collect.CronJob{Schedule: "30 8 * * *", Command: "/bin/true", Line: 2})
	f.When.Value = "09:15"
	f.Command.Value = "/bin/false"
	if !f.Changed() {
		t.Fatal("edit not detected")
	}
	f.Reset()
	if f.Changed() || f.When.Value != "08:30" || f.Command.Value != "/bin/true" || f.EditLine != 2 {
		t.Fatalf("reset: %+v", f)
	}
}

func TestUpdateRewritesLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "root")
	body := "# comment\n*/5 * * * * /bin/true\n\n0 3 * * * /bin/false\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	oldPath, oldReload := cronPath, cronReload
	cronPath = func() string { return path }
	reloaded := 0
	cronReload = func() error { reloaded++; return nil }
	t.Cleanup(func() { cronPath, cronReload = oldPath, oldReload })

	if err := Update(3, "0 3 * * * /bin/false", "30 4 * * *", "/bin/echo hi"); err != nil {
		t.Fatalf("update: %v", err)
	}
	raw, _ := os.ReadFile(path)
	want := "# comment\n*/5 * * * * /bin/true\n\n30 4 * * * /bin/echo hi\n"
	if string(raw) != want {
		t.Fatalf("file: %q", raw)
	}
	if reloaded != 1 {
		t.Fatalf("reload calls: %d", reloaded)
	}

	// a stale line number is reported instead of writing somewhere random
	if err := Update(9, "1 2 * * * /bin/true", "1 2 * * *", "/bin/true"); err == nil {
		t.Fatal("stale line accepted")
	}
	// so is a line that now holds a different job
	if err := Update(3, "0 3 * * * /bin/true", "1 2 * * *", "/bin/true"); err == nil {
		t.Fatal("changed line accepted")
	}
}

func TestWeekdayHiddenForIntervals(t *testing.T) {
	kit.SetLang("en")
	f := NewForm()
	if f.FieldRow(FocusWeekday) != -1 {
		t.Fatal("weekday row visible for an interval")
	}
	if got := len(f.Fields()); got != 3 {
		t.Fatalf("interval rows: %d", got)
	}

	// down from the value row skips the hidden weekday row
	f.Focus = FocusWhen
	f.StepFocus(+1)
	if f.Focus != FocusCommand {
		t.Fatalf("after when: %d", f.Focus)
	}
	f.StepFocus(+1)
	if f.Focus != FocusApply {
		t.Fatalf("after command: %d", f.Focus)
	}
	f.StepFocus(-1)
	if f.Focus != FocusCommand {
		t.Fatalf("back: %d", f.Focus)
	}

	// a fixed-time schedule keeps the row
	f.Kind.Index = KindDaily
	if f.FieldRow(FocusWeekday) != 2 {
		t.Fatalf("daily rows: %v", f.Fields())
	}
	f.Focus = FocusWhen
	f.StepFocus(+1)
	if f.Focus != FocusWeekday {
		t.Fatalf("daily down: %d", f.Focus)
	}
}

func TestSetEnabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "root")
	body := "#OFF# */5 * * * * /bin/true\n0 3 * * * /bin/false\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	oldPath, oldReload := cronPath, cronReload
	cronPath = func() string { return path }
	reloaded := 0
	cronReload = func() error { reloaded++; return nil }
	t.Cleanup(func() { cronPath, cronReload = oldPath, oldReload })

	job := collect.CronJob{Schedule: "*/5 * * * *", Command: "/bin/true", Line: 0}
	if err := SetEnabled(job, true); err != nil {
		t.Fatalf("enable: %v", err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "*/5 * * * * /bin/true\n0 3 * * * /bin/false\n" {
		t.Fatalf("enabled file: %q", raw)
	}

	if err := SetEnabled(job, false); err != nil {
		t.Fatalf("disable: %v", err)
	}
	raw, _ = os.ReadFile(path)
	if string(raw) != "#OFF# */5 * * * * /bin/true\n0 3 * * * /bin/false\n" {
		t.Fatalf("disabled file: %q", raw)
	}
	if reloaded != 2 {
		t.Fatalf("reload calls: %d", reloaded)
	}

	// a line that no longer holds this job is refused
	stale := collect.CronJob{Schedule: "*/5 * * * *", Command: "/bin/other", Line: 0}
	if err := SetEnabled(stale, true); err == nil {
		t.Fatal("stale job accepted")
	}
}
