package collect

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// readWifi reports the configured radios. Hardware presence comes from
// mac80211's phy devices; a device with a wifi-device section but no phy still
// counts so the settings stay editable while the radio is down.
func readWifi() WifiInfo {
	secs := parseUCI("/etc/config/wireless")
	phys, _ := filepath.Glob("/sys/class/ieee80211/phy*")

	var radios []uciSection
	for _, s := range secs {
		if s.typ == "wifi-device" {
			radios = append(radios, s)
		}
	}
	if len(phys) == 0 && len(radios) == 0 {
		return WifiInfo{}
	}

	w := WifiInfo{Supported: true}
	seen := map[string]bool{}
	for _, r := range radios {
		band := bandLabel(r)
		if seen[band] {
			continue // one tab per band; a second radio on the same band is not editable here
		}
		seen[band] = true

		b := WifiBand{
			Name:    band,
			Radio:   r.name,
			Enabled: r.opts["disabled"] != "1",
			Channel: channelRaw(r),
			Width:   widthMHz(r.opts["htmode"]),
		}
		if ap, ok := firstAP(secs, r.name); ok {
			b.Iface = ap.name
			b.SSID = ap.opts["ssid"]
			b.Key = ap.opts["key"]
			b.Security = securityRaw(ap.opts["encryption"])
			b.Enabled = b.Enabled && ap.opts["disabled"] != "1"
		}
		w.Bands = append(w.Bands, b)
	}
	sort.SliceStable(w.Bands, func(i, j int) bool { return bandRank(w.Bands[i].Name) < bandRank(w.Bands[j].Name) })
	return w
}

// firstAP returns the first AP interface on a radio; that is the one the edit
// form writes to.
func firstAP(secs []uciSection, radio string) (uciSection, bool) {
	for _, s := range secs {
		if s.typ != "wifi-iface" || s.opts["device"] != radio {
			continue
		}
		if mode := s.opts["mode"]; mode != "" && mode != "ap" {
			continue
		}
		return s, true
	}
	return uciSection{}, false
}

// channelRaw is the value uci stores, ready to be edited: uci uses "auto" for
// unset/0 channels.
func channelRaw(radio uciSection) string {
	ch := strings.TrimSpace(radio.opts["channel"])
	if ch == "" || ch == "0" {
		return "auto"
	}
	return ch
}

// securityRaw strips the cipher suffix (psk2+ccmp -> psk2) so the editor can
// cycle through the encryption modes it knows.
func securityRaw(enc string) string {
	return strings.SplitN(enc, "+", 2)[0]
}

func bandRank(band string) int {
	switch band {
	case "2.4 GHz":
		return 0
	case "5 GHz":
		return 1
	case "6 GHz":
		return 2
	}
	return 3
}

// bandLabel follows the modern `band` option first, then the legacy hwmode,
// then the even older `bands` list.
func bandLabel(radio uciSection) string {
	switch radio.opts["band"] {
	case "2g":
		return "2.4 GHz"
	case "5g":
		return "5 GHz"
	case "6g":
		return "6 GHz"
	}
	switch radio.opts["hwmode"] {
	case "11a", "11na", "11ac":
		return "5 GHz"
	case "11b", "11g", "11ng":
		return "2.4 GHz"
	case "11ad":
		return "60 GHz"
	}
	if strings.Contains(radio.opts["bands"], "5g") {
		return "5 GHz"
	}
	if strings.Contains(radio.opts["bands"], "2g") {
		return "2.4 GHz"
	}
	return ""
}

// widthMHz reads the trailing number of HT20/HT40+/VHT80/HE160/EHT320.
func widthMHz(htmode string) int {
	htmode = strings.TrimRight(htmode, "+-")
	i := len(htmode)
	for i > 0 && htmode[i-1] >= '0' && htmode[i-1] <= '9' {
		i--
	}
	if i == len(htmode) {
		return 0
	}
	n, err := strconv.Atoi(htmode[i:])
	if err != nil || n < 20 {
		return 0
	}
	return n
}
