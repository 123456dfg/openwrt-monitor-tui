// Mouse event contracts. Components own their hit-testing and their state
// changes: a pane translates screen coordinates into component coordinates,
// hands the event over, and then orchestrates whatever the state change means
// (a command, a note, a focus move). Purely drawn components — labels,
// charts — have no mouse contract at all.
package components

// Clickable is a component that reacts to a pointer press at (x, y), in cells
// relative to the component's own top-left corner. Click mutates the
// component's state and reports whether it consumed the event.
type Clickable interface {
	Click(x, y int) bool
}

// Wheelable is a component that reacts to a wheel step; d is -1 for wheel-up
// and +1 for wheel-down.
type Wheelable interface {
	Wheel(d int) bool
}
