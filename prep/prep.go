// Package prep runs before any other package's init (its import path sorts
// first) and disables the terminal background-color probe that Bubble Tea v1
// performs in its own init().
//
// Without this, starting the TUI on a terminal that does not answer the OSC 11
// query (PuTTY, busybox serial console, ...) blocks for termenv's 5s OSC
// timeout before the first frame is drawn.
package prep

import "github.com/charmbracelet/lipgloss"

func init() {
	// Pre-seed the answer so the probe is skipped entirely. Every colour used
	// by this UI is a fixed ANSI colour, so the detected value does not matter.
	lipgloss.SetHasDarkBackground(true)
}
