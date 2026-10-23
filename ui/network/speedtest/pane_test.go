package speedtest

import (
	"strings"
	"testing"

	"openwrt-monitor-tui/ui/kit"
)

func TestPaneIdle(t *testing.T) {
	kit.SetLang("en")
	out := strings.Join(Pane(New()), "\n")
	if !strings.Contains(out, "[ Start ]") {
		t.Fatalf("bracketed button missing:\n%s", out)
	}
	if strings.Contains(out, "|>") {
		t.Fatal("not-entered pane shows the play-style row")
	}
	if strings.Contains(out, "testing") {
		t.Fatal("idle pane shows a run in progress")
	}
}

// Entering the sub-tab turns the bracketed button into the play-style row.
func TestPaneEntered(t *testing.T) {
	kit.SetLang("en")
	st := New()
	st.Active = true
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, kit.Glyph.Sel+" "+kit.Tr.SpeedStart) || strings.Contains(out, "[ Start ]") {
		t.Fatalf("play-style row missing:\n%s", out)
	}

	// While a run is in flight the label swaps, the form stays.
	st.Running = true
	out = strings.Join(Pane(st), "\n")
	if !strings.Contains(out, kit.Glyph.Sel+" "+kit.Tr.SpeedRunning) {
		t.Fatalf("running label missing:\n%s", out)
	}
}

func TestPaneLive(t *testing.T) {
	kit.SetLang("en")
	st := New()
	st.Phase = PhaseDown
	st.Live = &Live{}
	st.Live.Down.Store(10e6 / 8)
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, "10.0 Mbps") || !strings.Contains(out, "testing") {
		t.Fatalf("live line missing:\n%s", out)
	}
}

func TestPaneServer(t *testing.T) {
	kit.SetLang("zh")
	st := New()
	st.Server = "Hong Kong, Hong Kong"
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, "服务器 Hong Kong, Hong Kong") {
		t.Fatalf("server region missing:\n%s", out)
	}
}

func TestPaneResult(t *testing.T) {
	kit.SetLang("en")
	st := New()
	st.Tool = "speedtest-go"
	st.Down = Rate{BPS: 10e6 / 8, Bytes: 12_500_000} // 10 Mbps
	st.Up = Rate{BPS: 3e6 / 8, Bytes: 3_750_000}     // 3 Mbps
	out := strings.Join(Pane(st), "\n")
	if !strings.Contains(out, "10.0 Mbps") || !strings.Contains(out, "3.0 Mbps") {
		t.Fatalf("rates missing:\n%s", out)
	}
	if !strings.Contains(out, "speedtest-go") {
		t.Fatalf("tool missing:\n%s", out)
	}
}

func TestStartClick(t *testing.T) {
	kit.SetLang("en")
	btns := StartButton()
	if !btns.Click(2, 0) || btns.Focus != 0 {
		t.Fatalf("start click: focus=%d", btns.Focus)
	}
	// Far outside the button nothing is hit.
	btns = StartButton()
	if btns.Click(60, 0) {
		t.Fatal("click beside the start button was consumed")
	}
}
