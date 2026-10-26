package system

import (
	"context"
	"errors"
	"openwrt-monitor-tui/ui/i18n"
	"strings"
	"testing"

	"openwrt-monitor-tui/ui/style"
)

func TestRunActions(t *testing.T) {
	var got []string
	old := execCmd
	execCmd = func(ctx context.Context, name string, args ...string) error {
		got = append(got, strings.TrimSpace(name+" "+strings.Join(args, " ")))
		return nil
	}
	t.Cleanup(func() { execCmd = old })

	for _, tc := range []struct {
		action Action
		want   []string
	}{
		{Reboot, []string{"reboot"}},
		{Poweroff, []string{"poweroff"}},
		{Factory, []string{"firstboot -y", "reboot"}},
	} {
		got = nil
		if err := Run(tc.action); err != nil {
			t.Fatalf("%v: %v", tc.action, err)
		}
		if strings.Join(got, "|") != strings.Join(tc.want, "|") {
			t.Fatalf("%v: %v", tc.action, got)
		}
	}
}

func TestFactoryStopsOnFirstbootError(t *testing.T) {
	old := execCmd
	calls := 0
	execCmd = func(ctx context.Context, name string, args ...string) error {
		calls++
		return errors.New("firstboot: denied")
	}
	t.Cleanup(func() { execCmd = old })

	if err := Run(Factory); err == nil {
		t.Fatal("error swallowed")
	}
	if calls != 1 {
		t.Fatalf("reboot ran after a failed reset: %d calls", calls)
	}
}

func TestPaneStates(t *testing.T) {
	i18n.SetLang("en")
	style.SetColor(false)
	t.Cleanup(func() { style.SetColor(true); i18n.SetLang("en") })

	lines := Pane(State{Active: true, Sel: 1, Confirm: true})
	if len(lines) <= ConfirmRow() {
		t.Fatalf("pane too short: %d", len(lines))
	}
	if !strings.Contains(lines[LineOffset()+1], style.Tr.SysPoweroff) {
		t.Fatalf("power off row missing: %q", lines[LineOffset()+1])
	}
	if !strings.Contains(lines[ConfirmRow()], style.Tr.SysPoweroff) {
		t.Fatalf("confirm line: %q", lines[ConfirmRow()])
	}

	busy := Pane(State{Busy: true})
	if !strings.Contains(busy[ConfirmRow()], style.Tr.SysRunning) {
		t.Fatalf("busy line: %q", busy[ConfirmRow()])
	}

	done := Pane(State{Note: style.Tr.SysDone})
	if !strings.Contains(strings.Join(done, "\n"), style.Tr.SysDone) {
		t.Fatal("note missing")
	}
	if got := Pane(State{}); got[ConfirmRow()] != "" {
		t.Fatalf("confirm line should be empty when idle: %q", got[ConfirmRow()])
	}
}
