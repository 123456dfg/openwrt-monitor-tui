package network

import (
	"openwrt-monitor-tui/ui/i18n"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/network/portscan"
	"openwrt-monitor-tui/ui/network/speedtest"
	"openwrt-monitor-tui/ui/style"
)

func TestTracker(t *testing.T) {
	i18n.SetLang("en")
	tr := NewTracker()
	for i := 0; i < 20; i++ {
		tr.Add([]collect.Iface{{Name: "wan", RxRate: float64(i * 10), TxRate: 5}})
	}
	c := tr.Chart("wan")
	if lipgloss.Width(c) != chartW {
		t.Fatalf("chart width %d", lipgloss.Width(c))
	}
	for _, r := range c {
		if r < 0x2800 || r > 0x28FF {
			t.Fatalf("chart not braille: %U", r)
		}
	}
}

func TestBodySpacing(t *testing.T) {
	i18n.SetLang("en")
	s := collect.Snapshot{At: time.Now(), Ifaces: []collect.Iface{
		{Name: "br-lan", Up: true, Addrs: []string{"192.168.1.1"}},
		{Name: "wan", Up: true, Addrs: []string{"10.0.0.2"}},
	}}
	body := strings.Join(Body(s, NewTracker(), 80, TabOverview, TabOverview, true, speedtest.New(), portscan.New()), "\n")
	if !strings.Contains(body, "br-lan") || !strings.Contains(body, "wan") {
		t.Fatal("interfaces missing")
	}
	// The sidebar lists both sub-tabs, settings style.
	if !strings.Contains(body, "Overview") || !strings.Contains(body, "Speed test") {
		t.Fatalf("sidebar missing:\n%s", body)
	}
	// One separator line between the interface rows (in the two-column join
	// the blank content row coincides with a sidebar row, so the separator
	// line carries only sidebar text).
	lines := strings.Split(body, "\n")
	br := -1
	for i, ln := range lines {
		if strings.Contains(strings.TrimSpace(ln), "br-lan ") {
			br = i
			break
		}
	}
	if br < 0 {
		t.Fatalf("br-lan row missing:\n%s", body)
	}
	if !strings.Contains(strings.TrimSpace(lines[br+2]), "wan ") {
		t.Fatalf("expected one separator line before the wan row:\n%s", body)
	}
}

func TestBodySpeedPane(t *testing.T) {
	i18n.SetLang("en")
	body := strings.Join(Body(collect.Snapshot{}, NewTracker(), 80, TabSpeed, TabSpeed, true, speedtest.New(), portscan.New()), "\n")
	if !strings.Contains(body, "Start") {
		t.Fatalf("start button missing:\n%s", body)
	}
}

// The sidebar marks three states: entered (solid triangle), preselected but
// not entered (dim dot) and untouched (no marker).
func TestSidebarEnterMarker(t *testing.T) {
	i18n.SetLang("en")
	t.Cleanup(func() { style.SetUnicode() })
	style.SetUnicode()

	entered := strings.Join(Body(collect.Snapshot{}, NewTracker(), 80, TabSpeed, TabSpeed, true, speedtest.New(), portscan.New()), "\n")
	if !strings.Contains(entered, style.Glyph.Sel+" Speed test") {
		t.Fatalf("entered marker missing:\n%s", entered)
	}
	if strings.Contains(entered, style.Glyph.Dot+" Speed test") {
		t.Fatalf("entered row still shows the preview dot:\n%s", entered)
	}

	preview := strings.Join(Body(collect.Snapshot{}, NewTracker(), 80, TabSpeed, TabSpeed, false, speedtest.New(), portscan.New()), "\n")
	if !strings.Contains(preview, style.Glyph.Dot+" Speed test") {
		t.Fatalf("preview dot missing:\n%s", preview)
	}
	if strings.Contains(preview, style.Glyph.Sel) {
		t.Fatalf("preview row shows the entered triangle:\n%s", preview)
	}

	idle := strings.Join(Body(collect.Snapshot{}, NewTracker(), 80, TabOverview, TabOverview, false, speedtest.New(), portscan.New()), "\n")
	if strings.Contains(idle, style.Glyph.Sel+" Speed test") || strings.Contains(idle, style.Glyph.Dot+" Speed test") {
		t.Fatalf("unselected speed row carries a marker:\n%s", idle)
	}
}
