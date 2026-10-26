// Pane is the Speed test tab content: a start action and the live and final
// results. Row indices here are shared with the mouse layer.
package speedtest

import (
	"fmt"

	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// Pane row indices, relative to the first line of Pane's output — which the
// Network tab aligns with the top of its sidebar. The mouse layer subtracts
// that offset before comparing.
const StartRow = 0

// State is everything the pane shows.
type State struct {
	Phase   int    // PhaseIdle / PhaseDown / PhaseUp / PhaseDone
	Running bool   // a measurement is in flight
	Active  bool   // the speed sub-tab has been entered: play-style action row
	Tool    string // what ran the test, e.g. "speedtest-go"
	Server  string // selected test server, e.g. "Hong Kong, Hong Kong"
	Down    Rate   // finished download result
	Up      Rate   // finished upload result

	Native any   // library-backed handle (speedtest-go server)
	Live   *Live // running averages polled by the pane, nil when idle

	Err  string
	Note string
}

// New builds the initial pane state.
func New() State { return State{} }

// Pane renders the speed test content.
func Pane(st State) []string {
	out := make([]string, 0, 8)
	out = append(out, startRow(st), "")
	if st.Server != "" {
		out = append(out, "  "+style.Dim.Render(style.Tr.SpeedServer+" "+st.Server))
	}
	switch {
	case st.Phase == PhaseDown:
		out = append(out, liveLine(st, false))
	case st.Phase == PhaseUp:
		out = append(out, liveLine(st, true))
	}
	if st.Note != "" {
		out = append(out, "  "+style.Good.Render(st.Note))
	}
	if st.Err != "" {
		out = append(out, "  "+style.Bad.Render(style.Tr.SpeedErr+": "+st.Err))
	}
	if st.Down.BPS > 0 || st.Up.BPS > 0 {
		out = append(out, "")
		out = append(out, finalLines(st)...)
	}
	return out
}

// liveLine shows the direction currently measured with its running average —
// the pane re-renders on every event, so the number moves in real time.
func liveLine(st State, upload bool) string {
	arrow, label := "↓", style.Tr.SpeedDown
	if upload {
		arrow, label = "↑", style.Tr.SpeedUp
	}
	var v float64
	if st.Live != nil {
		if upload {
			v = float64(st.Live.Up.Load())
		} else {
			v = float64(st.Live.Down.Load())
		}
	}
	line := arrow + " " + label + " " + fmt.Sprintf("%.1f Mbps", v*8/1e6)
	return "  " + style.Dim.Render(style.Tr.SpeedRunning+"   "+line)
}

// finalLines renders the tool line and the two result rows; speedtest-go
// always measures both directions.
func finalLines(st State) []string {
	out := []string{"  " + style.Dim.Render(style.Tr.SpeedTool+" "+st.Tool)}
	down := fmt.Sprintf("↓ %s %.1f Mbps  (%s)",
		style.Tr.SpeedDown, st.Down.BPS*8/1e6, style.HumanBytes(uint64(st.Down.Bytes)))
	up := fmt.Sprintf("↑ %s %.1f Mbps  (%s)",
		style.Tr.SpeedUp, st.Up.BPS*8/1e6, style.HumanBytes(uint64(st.Up.Bytes)))
	return append(out, "  "+down, "  "+up)
}

// StartButton is the action row in its idle, not-entered form; the empty
// second label hides it.
func StartButton() components.Buttons {
	b := components.Buttons{Labels: [2]string{style.Tr.SpeedStart, ""}}
	b.Focus = -1
	return b
}

// startRow renders the action row. Once the sub-tab has been entered the
// bracketed button turns into a play-style row (▸ 开始测速), so the user can
// tell the pane is live; before that it stays the plain [ 开始测速 ] button.
func startRow(st State) string {
	if st.Active {
		label := style.Tr.SpeedStart
		row := style.Accent
		if st.Running {
			label, row = style.Tr.SpeedRunning, style.Dim
		}
		return "  " + row.Render(style.Glyph.Sel+" "+label)
	}
	b := StartButton()
	if st.Running {
		b.Busy, b.BusyLabel = true, style.Tr.SpeedRunning
	}
	return "  " + b.View()
}
