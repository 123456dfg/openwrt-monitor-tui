package system

import (
	"fmt"

	"openwrt-monitor-tui/ui/style"
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
		return style.Tr.SysPoweroff
	case Factory:
		return style.Tr.SysFactory
	}
	return style.Tr.SysReboot
}

func Desc(a Action) string {
	switch a {
	case Poweroff:
		return style.Tr.SysPoweroffDesc
	case Factory:
		return style.Tr.SysFactoryDesc
	}
	return style.Tr.SysRebootDesc
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
	out := []string{"", "  " + style.Dim.Render(style.Tr.SysHint), ""}
	for i := 0; i < int(Count); i++ {
		marker := "  "
		row := style.Label
		if st.Active && i == st.Sel {
			marker, row = style.Glyph.Sel+" ", style.Accent
		}
		line := "  " + marker + row.Render(style.PadRight(Label(Action(i)), 14)) +
			"  " + style.Dim.Render(Desc(Action(i)))
		if st.Confirm && i == st.Sel {
			line = "  " + marker + style.Bad.Render(style.PadRight(Label(Action(i)), 14)) +
				"  " + style.Dim.Render(Desc(Action(i)))
		}
		out = append(out, line)
	}
	out = append(out, "", confirmLine(st))
	if st.Note != "" {
		out = append(out, "  "+style.Good.Render(st.Note))
	}
	if st.Err != "" {
		out = append(out, "  "+style.Bad.Render(st.Err))
	}
	return out
}

func confirmLine(st State) string {
	switch {
	case st.Busy:
		return "  " + style.Warn.Render(style.Tr.SysRunning)
	case st.Confirm:
		return "  " + style.Bad.Render(fmt.Sprintf(style.Tr.SysConfirm, Label(Action(st.Sel))))
	}
	return ""
}

// LineOffset is the body line index of the first action row.
func LineOffset() int { return 3 }

// ConfirmRow is the body line index of the confirmation line.
func ConfirmRow() int { return LineOffset() + int(Count) + 1 }
