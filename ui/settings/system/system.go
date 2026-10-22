package system

import (
	"fmt"

	"openwrt-monitor-tui/ui/kit"
)

// Action is one entry of the system page.
type Action int

const (
	Reboot Action = iota
	Poweroff
	Factory
	Count
)

func Label(a Action) string {
	switch a {
	case Poweroff:
		return kit.Tr.SysPoweroff
	case Factory:
		return kit.Tr.SysFactory
	}
	return kit.Tr.SysReboot
}

func Desc(a Action) string {
	switch a {
	case Poweroff:
		return kit.Tr.SysPoweroffDesc
	case Factory:
		return kit.Tr.SysFactoryDesc
	}
	return kit.Tr.SysRebootDesc
}

// State is everything the system page draws.
type State struct {
	Sel     int
	Confirm bool
	Busy    bool
	Active  bool
	Note    string
	Err     string
}

// Pane lists the actions and reserves a line for the confirmation. Messages go
// below it, so the row positions never shift.
func Pane(st State) []string {
	out := []string{"", "  " + kit.Dim.Render(kit.Tr.SysHint), ""}
	for i := 0; i < int(Count); i++ {
		marker := "  "
		style := kit.Label
		if st.Active && i == st.Sel {
			marker, style = kit.Glyph.Sel+" ", kit.Accent
		}
		line := "  " + marker + style.Render(kit.PadRight(Label(Action(i)), 14)) +
			"  " + kit.Dim.Render(Desc(Action(i)))
		if st.Confirm && i == st.Sel {
			line = "  " + marker + kit.Bad.Render(kit.PadRight(Label(Action(i)), 14)) +
				"  " + kit.Dim.Render(Desc(Action(i)))
		}
		out = append(out, line)
	}
	out = append(out, "", confirmLine(st))
	if st.Note != "" {
		out = append(out, "  "+kit.Good.Render(st.Note))
	}
	if st.Err != "" {
		out = append(out, "  "+kit.Bad.Render(st.Err))
	}
	return out
}

func confirmLine(st State) string {
	switch {
	case st.Busy:
		return "  " + kit.Warn.Render(kit.Tr.SysRunning)
	case st.Confirm:
		return "  " + kit.Bad.Render(fmt.Sprintf(kit.Tr.SysConfirm, Label(Action(st.Sel))))
	}
	return ""
}

// LineOffset is the body line index of the first action row.
func LineOffset() int { return 3 }

// ConfirmRow is the body line index of the confirmation line.
func ConfirmRow() int { return LineOffset() + int(Count) + 1 }
