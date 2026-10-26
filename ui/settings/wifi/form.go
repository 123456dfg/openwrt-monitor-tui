package wifi

import (
	"strconv"
	"strings"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/style"
	"openwrt-monitor-tui/ui/widgets/components"
)

// Focus order of the Wi-Fi form; Apply/Cancel are reached like any field.
const (
	FocusEnabled = iota
	FocusSSID
	FocusPassword
	FocusChannel
	FocusSecurity
	FocusApply
	FocusCancel
	FocusMax
)

// Form is the editable copy of one band. The Org* fields are the values last
// read from /etc/config/wireless: they drive the changed check and Cancel.
// Each row is a widget, so the same controls can be reused by other pages.
type Form struct {
	Loaded bool
	Band   string // band the values belong to
	Radio  string // wifi-device section name
	Iface  string // wifi-iface section name, "" when the radio has no AP

	Enabled  components.Checkbox
	SSID     components.Input
	Key      components.Input
	Channel  components.Select
	Security components.Select
	Width    int

	OrgEnabled  bool
	OrgSSID     string
	OrgKey      string
	OrgChannel  string
	OrgSecurity string

	Focus   int
	Editing bool
	Busy    bool
}

func FromBand(b collect.WifiBand, focus int) Form {
	return Form{
		Loaded: true,
		Band:   b.Name,
		Radio:  b.Radio,
		Iface:  b.Iface,

		Enabled:  components.Checkbox{On: b.Enabled},
		SSID:     components.NewInput(b.SSID, 32),
		Key:      components.NewInput(b.Key, 63),
		Channel:  components.NewSelect(channelOptions(b.Name, b.Channel), b.Channel),
		Security: components.NewSelect(securityOptions(b.Security), b.Security),
		Width:    b.Width,

		OrgEnabled:  b.Enabled,
		OrgSSID:     b.SSID,
		OrgKey:      b.Key,
		OrgChannel:  b.Channel,
		OrgSecurity: b.Security,

		Focus: focus,
	}
}

func (f Form) Changed() bool {
	return f.Enabled.On != f.OrgEnabled || f.SSID.Value != f.OrgSSID ||
		f.Key.Value != f.OrgKey || f.Channel.Value() != f.OrgChannel ||
		f.Security.Value() != f.OrgSecurity
}

// Reset restores the values read from disk.
func (f *Form) Reset() {
	f.Enabled.On = f.OrgEnabled
	f.SSID.Value = f.OrgSSID
	f.Key.Value = f.OrgKey
	f.Channel.Set(f.OrgChannel)
	f.Security.Set(f.OrgSecurity)
	f.Editing = false
}

// Adopt makes the current values the new baseline after a successful apply.
func (f *Form) Adopt() {
	f.OrgEnabled = f.Enabled.On
	f.OrgSSID = f.SSID.Value
	f.OrgKey = f.Key.Value
	f.OrgChannel = f.Channel.Value()
	f.OrgSecurity = f.Security.Value()
}

func (f Form) Validate() string {
	ch := strings.TrimSpace(f.Channel.Value())
	if ch != "auto" {
		n, err := strconv.Atoi(ch)
		if err != nil || n < 1 || n > 233 {
			return style.Tr.ValChannel
		}
	}
	if f.Iface == "" {
		return ""
	}
	if f.SSID.Value == "" {
		return style.Tr.ValSSIDEmpty
	}
	if len(f.SSID.Value) > 32 {
		return style.Tr.ValSSIDLong
	}
	if NeedsKey(f.Security.Value()) && (len(f.Key.Value) < 8 || len(f.Key.Value) > 63) {
		return style.Tr.ValPassword
	}
	return ""
}

// Plan turns the edits into uci writes; untouched options are not rewritten.
func (f Form) Plan() []Set {
	var out []Set
	if f.Enabled.On != f.OrgEnabled {
		// Enabled covers both sections, so turn the AP on/off with the radio;
		// otherwise an iface left disabled would silently flip the switch back.
		out = append(out, Set{f.Radio, "disabled", boolStr(!f.Enabled.On)})
		if f.Iface != "" {
			out = append(out, Set{f.Iface, "disabled", boolStr(!f.Enabled.On)})
		}
	}
	if ch := f.Channel.Value(); ch != f.OrgChannel {
		out = append(out, Set{f.Radio, "channel", strings.TrimSpace(ch)})
	}
	if f.Iface == "" {
		return out
	}
	if f.SSID.Value != f.OrgSSID {
		out = append(out, Set{f.Iface, "ssid", f.SSID.Value})
	}
	if sec := f.Security.Value(); sec != f.OrgSecurity {
		out = append(out, Set{f.Iface, "encryption", sec})
	}
	// Dropping the key when the network becomes open is deliberate: otherwise
	// the old passphrase stays in the config and comes back on the next WPA2.
	if f.Key.Value != f.OrgKey || NeedsKey(f.Security.Value()) != NeedsKey(f.OrgSecurity) {
		key := f.Key.Value
		if !NeedsKey(f.Security.Value()) {
			key = ""
		}
		out = append(out, Set{f.Iface, "key", key})
	}
	return out
}

// Text is the input widget of the focused row, or nil when the row is not a
// text field.
func (f *Form) Text() *components.Input {
	switch f.Focus {
	case FocusSSID:
		return &f.SSID
	case FocusPassword:
		return &f.Key
	}
	return nil
}

// Picker is the select widget of the focused row, or nil.
func (f *Form) Picker() *components.Select {
	switch f.Focus {
	case FocusChannel:
		return &f.Channel
	case FocusSecurity:
		return &f.Security
	}
	return nil
}

// FocusValue is the current value of the focused row, typed or picked.
func (f Form) FocusValue() string {
	if p := f.Picker(); p != nil {
		return p.Value()
	}
	if in := f.Text(); in != nil {
		return in.Value
	}
	return ""
}

// CommitEdit closes the row, trimming the SSID the way uci does. Passwords may
// legitimately contain spaces, so they are left alone.
func (f *Form) CommitEdit() {
	if f.Focus == FocusSSID {
		f.SSID.Value = strings.TrimSpace(f.SSID.Value)
	}
	f.Editing = false
}

func NeedsKey(security string) bool {
	switch security {
	case "", "none", "owe":
		return false
	}
	return true
}

func boolStr(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

var securityCycle = []string{"none", "psk2", "psk-mixed", "sae", "sae-mixed", "owe"}

// securityOptions is the cycle order. A value the list does not know (an
// unset option, an EAP mode) is kept as its own entry so the form loads it
// unchanged.
func securityOptions(cur string) []string {
	opts := append([]string(nil), securityCycle...)
	for _, o := range opts {
		if o == cur {
			return opts
		}
	}
	return append(opts, cur)
}

// channelsFor lists the standard 20 MHz channel numbers per band. 144 and the
// DFS block are included for 5 GHz; a country may allow fewer.
func channelsFor(band string) []int {
	switch band {
	case "2.4 GHz":
		out := make([]int, 0, 13)
		for ch := 1; ch <= 13; ch++ {
			out = append(out, ch)
		}
		return out
	case "5 GHz":
		return []int{36, 40, 44, 48, 52, 56, 60, 64, 100, 104, 108, 112,
			116, 120, 124, 128, 132, 136, 140, 144, 149, 153, 157, 161, 165}
	case "6 GHz":
		out := make([]int, 0, 59)
		for ch := 1; ch <= 233; ch += 4 {
			out = append(out, ch)
		}
		return out
	case "60 GHz":
		return []int{1, 2, 3, 4, 5, 6}
	}
	return nil
}

// channelOptions is "auto" followed by the band's channels. A channel already
// in the config but outside the list (e.g. 14) is kept so it can be left as-is.
func channelOptions(band, cur string) []string {
	opts := append([]string{"auto"}, intStrings(channelsFor(band))...)
	for _, o := range opts {
		if o == cur {
			return opts
		}
	}
	if cur != "" && cur != "auto" {
		opts = append(opts, cur)
	}
	return opts
}

func intStrings(ns []int) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = strconv.Itoa(n)
	}
	return out
}

func ChannelHint(band string) string {
	switch band {
	case "6 GHz":
		return style.Tr.Hint6
	case "60 GHz":
		return style.Tr.Hint60
	case "2.4 GHz":
		return style.Tr.Hint24
	case "5 GHz":
		return style.Tr.Hint5
	}
	return style.Tr.HintOther
}

func SecurityLabel(enc string) string {
	switch strings.SplitN(enc, "+", 2)[0] {
	case "", "none":
		return style.Tr.Open
	case "psk":
		return "WPA-PSK"
	case "psk2":
		return "WPA2-PSK"
	case "psk-mixed":
		return "WPA/WPA2-PSK"
	case "sae":
		return "WPA3-SAE"
	case "sae-mixed":
		return "WPA2/WPA3"
	case "owe":
		return "OWE"
	case "wpa":
		return "WPA-EAP"
	case "wpa2":
		return "WPA2-EAP"
	case "wpa-mixed":
		return "WPA/WPA2-EAP"
	}
	return enc
}
