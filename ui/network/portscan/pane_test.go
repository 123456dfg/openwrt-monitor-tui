package portscan

import (
	"openwrt-monitor-tui/ui/i18n"
	"strings"
	"testing"

	"openwrt-monitor-tui/task"
	"openwrt-monitor-tui/ui/style"
)

func sampleState(phase int, active bool) State {
	st := New()
	st.Phase, st.Active = phase, active
	return st
}

func TestPaneIdle(t *testing.T) {
	i18n.SetLang("en")
	out := strings.Join(Pane(New()), "\n")
	if !strings.Contains(out, "[ Start scan ]") {
		t.Fatalf("bracketed button missing:\n%s", out)
	}
	if strings.Contains(out, "|>") && !strings.Contains(out, style.Glyph.Sel) {
		t.Fatal("not-entered pane shows the play-style row")
	}
}

func TestPaneEntered(t *testing.T) {
	i18n.SetLang("en")
	st := sampleState(PhaseIdle, true)
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, style.Glyph.Sel+" "+style.Tr.PortStart) || strings.Contains(out, "[ Start scan ]") {
		t.Fatalf("play-style row missing:\n%s", out)
	}

	// While scanning, the live counter shows.
	st.Phase = PhaseRunning
	st.Progress = &task.ScanProgress{}
	st.Progress.Total.Store(240)
	st.Progress.Probed.Store(96)
	out = strings.Join(Pane(st), "\n")
	if !strings.Contains(out, "probed 96/240") {
		t.Fatalf("progress missing:\n%s", out)
	}
}

func TestPaneDone(t *testing.T) {
	i18n.SetLang("en")
	st := sampleState(PhaseDone, true)
	st.Hosts = []task.HostResult{
		{IP: "192.168.1.20", MAC: "aa:bb:cc:dd:ee:01", Ports: []int{22, 80}},
		{IP: "192.168.1.21", MAC: "aa:bb:cc:dd:ee:02"},
	}
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, "found 2 hosts, 2 open ports") {
		t.Fatalf("summary missing:\n%s", out)
	}
	if !strings.Contains(out, "192.168.1.20") || !strings.Contains(out, "22 80") {
		t.Fatalf("host row missing:\n%s", out)
	}
	if !strings.Contains(out, "aa:bb:cc:dd:ee:02 -") {
		t.Fatalf("portless host dash missing:\n%s", out)
	}
}

func TestPaneEmptyDone(t *testing.T) {
	i18n.SetLang("en")
	st := sampleState(PhaseDone, true)
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, "no open ports found") {
		t.Fatalf("empty hint missing:\n%s", out)
	}
}
