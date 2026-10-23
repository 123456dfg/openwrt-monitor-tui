package components

// Select is a fixed option list with one chosen entry. Begin/Step/Cancel give
// it the same edit cycle as Input, so a form can treat both the same way.
type Select struct {
	Options []string
	Index   int

	saved int
}

// NewSelect builds a select around cur. A value that is not in options (an
// older config, a channel the band list does not cover) is appended instead of
// being silently replaced, so loading a form never marks it dirty.
func NewSelect(options []string, cur string) Select {
	s := Select{Options: append([]string(nil), options...)}
	s.Index = s.indexOf(cur)
	return s
}

func (s Select) Value() string {
	if len(s.Options) == 0 {
		return ""
	}
	return s.Options[s.Index]
}

// Set picks v, appending it when the list does not have it yet.
func (s *Select) Set(v string) { s.Index = s.indexOf(v) }

func (s *Select) indexOf(v string) int {
	for i, o := range s.Options {
		if o == v {
			return i
		}
	}
	s.Options = append(s.Options, v)
	return len(s.Options) - 1
}

func (s *Select) Begin() { s.saved = s.Index }

func (s *Select) Step(d int) {
	if len(s.Options) == 0 {
		return
	}
	s.Index = ((s.Index+d)%len(s.Options) + len(s.Options)) % len(s.Options)
}

func (s *Select) Cancel() { s.Index = s.saved }

// Wheel moves the selection by d; a select is switched with the wheel or the
// arrows once its row has been entered — a click only takes focus, which is
// the pane's business.
func (s *Select) Wheel(d int) bool {
	s.Step(d)
	return true
}

// View draws the chosen entry; an editing select is bracketed and owns the
// arrow keys.
func (s Select) View(editing bool) string {
	v := s.Value()
	if editing {
		return "[" + v + "]"
	}
	return v
}
