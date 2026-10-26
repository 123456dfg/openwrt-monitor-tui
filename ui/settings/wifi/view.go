package wifi

import (
	"fmt"
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// State is everything the Wi-Fi pane needs to draw itself.
type State struct {
	Form     Form
	Bands    []collect.WifiBand
	Band     int
	Active   bool // content has the focus (sidebar not selected)
	ShowKeys bool
	Note     string
	Err      string
}

// Pane renders the band strip, messages and the form of the selected band.
func Pane(st State) []string {
	var out []string
	if len(st.Bands) > 0 {
		out = append(out, BandBar(st.Bands, st.Band), "")
	}
	if len(st.Bands) == 0 {
		return append(out, "  "+style.Dim.Render(style.Tr.NoRadio))
	}
	if st.Err != "" {
		out = append(out, "  "+style.Bad.Render(st.Err))
	}
	if st.Note != "" {
		out = append(out, "  "+style.Good.Render(st.Note))
	}
	if st.Err != "" || st.Note != "" {
		out = append(out, "")
	}
	return append(out, rows(st)...)
}

// FormRowOffset is the line index of the Enabled row inside the pane.
func FormRowOffset(st State) int {
	off := 0
	if len(st.Bands) > 0 {
		off += 2 // band strip + blank line
	}
	if st.Err != "" {
		off++
	}
	if st.Note != "" {
		off++
	}
	if st.Err != "" || st.Note != "" {
		off++
	}
	return off
}

func rows(st State) []string {
	f := st.Form
	focus := f.Focus
	if !st.Active {
		focus = -1
	}

	editing := func(i int) bool { return f.Editing && i == focus }

	pw := f.Key.View(editing(FocusPassword))
	if !st.ShowKeys && !editing(FocusPassword) {
		if f.Key.Value == "" {
			pw = "-"
		} else {
			pw = components.Mask(8) // fixed width: the real length is secret
		}
	}
	ch := f.Channel.View(editing(FocusChannel))
	if f.Width > 0 && !editing(FocusChannel) {
		ch += fmt.Sprintf("  (%d MHz)", f.Width)
	}

	fields := [][2]string{
		{style.Tr.Enabled, f.Enabled.String()},
		{style.Tr.SSID, f.SSID.View(editing(FocusSSID))},
		{style.Tr.Password, pw},
		{style.Tr.Channel, ch},
		{style.Tr.Security, SecurityLabel(f.Security.Value())},
	}

	out := make([]string, 0, len(fields)+3)
	for i, r := range fields {
		out = append(out, fieldRow(i, f, focus, r[0], r[1]))
	}
	out = append(out, "", buttonsRow(f, focus))
	return out
}

// buttonsRow renders Apply and Cancel side by side.
func buttonsRow(f Form, focus int) string {
	return actionRow(f, focus).View()
}

func actionRow(f Form, focus int) components.Buttons {
	return components.Buttons{
		Labels:    [2]string{style.Tr.Apply, style.Tr.Cancel},
		Focus:     buttonIndex(focus),
		Busy:      f.Busy,
		BusyLabel: style.Tr.Applying,
	}
}

func buttonIndex(focus int) int {
	switch focus {
	case FocusApply:
		return 0
	case FocusCancel:
		return 1
	}
	return -1
}

// ButtonHits was absorbed by components.Buttons.Click: the pane constructs
// the same Buttons (labels come from the translations) and lets the component
// resolve the press.

func fieldRow(i int, f Form, focus int, name, value string) string {
	marker := "  "
	nameStyle, valStyle := style.Label, style.Style()
	if i == focus {
		marker, nameStyle, valStyle = style.Glyph.Sel+" ", style.Accent, style.Accent
	}
	if (i == FocusSSID || i == FocusPassword || i == FocusSecurity) && f.Iface == "" {
		nameStyle, valStyle = style.Dim, style.Dim
	}
	editingHere := f.Editing && i == focus
	list := i == FocusChannel || i == FocusSecurity
	line := "  " + marker + nameStyle.Render(style.PadRight(name, 9)) + " " + valStyle.Render(value)
	if editingHere && list {
		line += "   " + style.Dim.Render("←/→")
	} else if i == FocusChannel && i == focus {
		line += "   " + style.Dim.Render(ChannelHint(f.Band))
	}
	return line
}

// BandBar is the tab strip inside the Wi-Fi pane. Only bands the device has
// are listed.
func BandBar(bands []collect.WifiBand, cur int) string {
	cur = min(cur, len(bands)-1)
	var b strings.Builder
	b.WriteString("  ")
	for i, band := range bands {
		if i > 0 {
			b.WriteString("   ")
		}
		label := ShortBand(band.Name)
		if i == cur {
			if style.ColorOn() {
				b.WriteString(style.Accent.Render(label))
			} else {
				b.WriteString("[" + label + "]")
			}
			continue
		}
		b.WriteString(style.Label.Render(label))
	}
	return b.String()
}

func ShortBand(band string) string {
	switch band {
	case "2.4 GHz":
		return "2.4G"
	case "5 GHz":
		return "5G"
	case "6 GHz":
		return "6G"
	case "":
		return "Wi-Fi"
	}
	return band
}

// BandHits mirrors BandBar's x range for click handling.
func BandHits(bands []collect.WifiBand, cur int) [][2]int {
	hits := make([][2]int, 0, len(bands))
	x := style.SidebarW + 4
	cur = min(cur, len(bands)-1)
	for i, band := range bands {
		w := len(ShortBand(band.Name))
		if i == cur && !style.ColorOn() {
			w += 2
		}
		hits = append(hits, [2]int{x, x + w - 1})
		x += w + 3
	}
	return hits
}
