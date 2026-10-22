package network

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
)

func TestTracker(t *testing.T) {
	kit.SetLang("en")
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
	kit.SetLang("en")
	s := collect.Snapshot{At: time.Now(), Ifaces: []collect.Iface{
		{Name: "br-lan", Up: true, Addrs: []string{"192.168.1.1"}},
		{Name: "wan", Up: true, Addrs: []string{"10.0.0.2"}},
	}}
	body := strings.Join(Body(s, NewTracker(), 80), "\n")
	if !strings.Contains(body, "br-lan") || !strings.Contains(body, "wan") {
		t.Fatal("interfaces missing")
	}
	if !strings.Contains(body, "\n\n  wan") {
		t.Fatalf("no blank line between interfaces:\n%s", body)
	}
}
