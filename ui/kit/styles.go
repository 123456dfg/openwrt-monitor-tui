package kit

import "github.com/charmbracelet/lipgloss"

// Palette used by every tab. SetColor replaces the whole set with plain
// styles so serial consoles do not pay for escape sequences.
var (
	Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	Label  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	Dim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	Good   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	Warn   = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Bad    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

// Glyphs are the box-drawing characters; SetASCII swaps them for serial
// consoles that cannot render them.
type Glyphs struct {
	BarFull, BarEmpty, Sep, Dot, Sel, Mask string
	Spark                                  string // eight levels, low to high
}

var Glyph = Glyphs{BarFull: "█", BarEmpty: "░", Sep: "─", Dot: "·", Sel: "▸", Mask: "•", Spark: "▁▂▃▄▅▆▇█"}

const (
	BarWidth = 20
	SidebarW = 16
)

// SetUnicode restores the default box-drawing glyphs. It is the counterpart of
// SetASCII, mainly useful in tests.
func SetUnicode() {
	asciiMode = false
	Glyph = Glyphs{BarFull: "█", BarEmpty: "░", Sep: "─", Dot: "·", Sel: "▸", Mask: "•", Spark: "▁▂▃▄▅▆▇█"}
}

func SetASCII() {
	asciiMode = true
	Glyph = Glyphs{BarFull: "#", BarEmpty: "-", Sep: "-", Dot: "|", Sel: ">", Mask: "*", Spark: " .:-=+*#"}
}

var (
	colorOn   = true
	asciiMode = false
)

// ColorOn reports whether colours are enabled; tabs fall back to bracket
// markers when they are not.
func ColorOn() bool { return colorOn }

// ASCII reports whether the ASCII glyph set is active. Charts fall back to bar
// glyphs because braille is not available there.
func ASCII() bool { return asciiMode }

func SetColor(on bool) {
	colorOn = on
	if on {
		Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		Label = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
		Dim = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		Good = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		Warn = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		Bad = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
		return
	}
	Accent, Label, Dim = lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle()
	Good, Warn, Bad = lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle()
}

// Style returns an unstyled base that callers can Render without colour.
func Style() lipgloss.Style { return lipgloss.NewStyle() }
