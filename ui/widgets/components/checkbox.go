package components

// Checkbox is a boolean switch rendered as [x] / [ ].
type Checkbox struct {
	On bool
}

func (c *Checkbox) Toggle() { c.On = !c.On }

// Click flips the switch; the whole row is the hit area.
func (c *Checkbox) Click(x, y int) bool {
	c.Toggle()
	return true
}

func (c Checkbox) String() string {
	if c.On {
		return "[x]"
	}
	return "[ ]"
}
