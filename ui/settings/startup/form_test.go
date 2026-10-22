package startup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
)

func info() collect.StartupInfo {
	return collect.StartupInfo{
		Available: true,
		Path:      "/etc/rc.local",
		Content: []string{
			"#!/bin/sh",
			"# a comment",
			"/bin/true",
			"#OFF# /bin/false",
			"",
			"exit 0",
		},
	}
}

func TestParseLines(t *testing.T) {
	f := FromInfo(info(), 0)
	if !f.Available || f.TaskCount() != 2 {
		t.Fatalf("tasks: %d", f.TaskCount())
	}
	first, _ := f.TaskAt(0)
	second, _ := f.TaskAt(1)
	if first.Text != "/bin/true" || !first.Enabled {
		t.Fatalf("first: %+v", first)
	}
	if second.Text != "/bin/false" || second.Enabled {
		t.Fatalf("second: %+v", second)
	}
	// hidden lines are preserved byte for byte
	body := f.Body()
	for _, want := range []string{"#!/bin/sh", "# a comment", "#OFF# /bin/false", "exit 0"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body lost %q:\n%s", want, body)
		}
	}
	if f.Changed() {
		t.Fatal("fresh form reported as changed")
	}
}

func TestToggleAndInsert(t *testing.T) {
	kit.SetLang("en")
	f := FromInfo(info(), 0)

	f.Toggle(0) // disable /bin/true
	if !f.Changed() {
		t.Fatal("toggle not detected")
	}
	if !strings.Contains(f.Body(), "#OFF# /bin/true") {
		t.Fatalf("body:\n%s", f.Body())
	}
	f.Toggle(0)
	if f.Changed() {
		t.Fatal("toggle back should be clean")
	}

	// a new task goes above the trailing exit
	f.Insert()
	if f.TaskCount() != 3 || f.Focus != 2 {
		t.Fatalf("tasks=%d focus=%d", f.TaskCount(), f.Focus)
	}
	f.Edit().Clear()
	for _, r := range "/bin/added" {
		f.Edit().Type(r)
	}
	f.CommitEdit()
	body := f.Body()
	if !strings.Contains(body, "/bin/added\nexit 0") {
		t.Fatalf("inserted in the wrong place:\n%s", body)
	}

	f.Delete(2)
	if f.TaskCount() != 2 {
		t.Fatalf("after delete: %d", f.TaskCount())
	}
}

func TestEditTask(t *testing.T) {
	kit.SetLang("en")
	f := FromInfo(info(), 0)
	if !f.Begin(1) {
		t.Fatal("begin failed")
	}
	f.Edit().Clear()
	for _, r := range "/bin/changed  " {
		f.Edit().Type(r)
	}
	f.CommitEdit()
	if f.Editing {
		t.Fatal("still editing")
	}
	if e, _ := f.TaskAt(1); e.Text != "/bin/changed" {
		t.Fatalf("text: %+v", e)
	}
}

func TestResetAndAdopt(t *testing.T) {
	kit.SetLang("en")
	f := FromInfo(info(), 0)
	f.Toggle(0)
	f.Reset()
	if f.Changed() {
		t.Fatal("reset did not restore")
	}
	f.Toggle(0)
	f.Adopt()
	if f.Changed() {
		t.Fatal("adopt did not rebase")
	}
	if !strings.Contains(f.Body(), "#OFF# /bin/true") {
		t.Fatalf("body after adopt:\n%s", f.Body())
	}
}

func TestSave(t *testing.T) {
	kit.SetLang("en")
	path := filepath.Join(t.TempDir(), "rc.local")
	if err := os.WriteFile(path, []byte("old\n"), 0o750); err != nil {
		t.Fatal(err)
	}
	old := startupFile
	startupFile = func() string { return path }
	t.Cleanup(func() { startupFile = old })

	f := FromInfo(info(), 0)
	if err := Save(f.Body()); err != nil {
		t.Fatalf("save: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != f.Body() {
		t.Fatalf("file: %q", raw)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o750 {
		t.Fatalf("mode: %v", st.Mode().Perm())
	}
	if err := Save(""); err != nil {
		t.Fatalf("empty save: %v", err)
	}
	if raw, _ := os.ReadFile(path); len(raw) != 0 {
		t.Fatalf("empty file: %q", raw)
	}
}
