package wifi

import (
	"openwrt-monitor-tui/ui/i18n"
	"testing"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/style"
)

func form() Form {
	return FromBand(collect.WifiBand{
		Name: "2.4 GHz", Radio: "radio0", Iface: "default_radio0",
		Enabled: true, SSID: "OpenWrt", Key: "secret123", Security: "psk2", Channel: "6",
	}, FocusEnabled)
}

func TestValidate(t *testing.T) {
	i18n.SetLang("en")
	f := form()
	if msg := f.Validate(); msg != "" {
		t.Fatalf("unexpected error: %s", msg)
	}
	f.SSID.Value = ""
	if f.Validate() != style.Tr.ValSSIDEmpty {
		t.Fatal("empty ssid accepted")
	}
	f = form()
	f.Key.Value = "short"
	if f.Validate() != style.Tr.ValPassword {
		t.Fatal("short key accepted")
	}
	f = form()
	f.Channel.Set("abc")
	if f.Validate() != style.Tr.ValChannel {
		t.Fatal("bad channel accepted")
	}
}

func TestChannelsAndSecurity(t *testing.T) {
	f := form()
	f.Channel.Step(+1)
	if f.Channel.Value() != "7" {
		t.Fatalf("channel %q", f.Channel.Value())
	}
	f.Channel.Set("13")
	f.Channel.Step(+1)
	if f.Channel.Value() != "auto" {
		t.Fatalf("wrap %q", f.Channel.Value())
	}
	f.Channel.Set("14") // kept because the config has it
	f.Channel.Step(+1)
	if f.Channel.Value() != "auto" {
		t.Fatalf("custom wrap %q", f.Channel.Value())
	}

	f = form()
	f.Security.Step(+1)
	if f.Security.Value() != "psk-mixed" {
		t.Fatalf("security %q", f.Security.Value())
	}
	// an unknown encryption mode survives loading and waits at the end
	odd := FromBand(collect.WifiBand{
		Name: "2.4 GHz", Radio: "radio0", Iface: "default_radio0", Enabled: true,
		SSID: "x", Key: "secret123", Security: "wpa2", Channel: "6",
	}, FocusEnabled)
	if odd.Security.Value() != "wpa2" || odd.Changed() {
		t.Fatalf("unknown security: %q changed=%v", odd.Security.Value(), odd.Changed())
	}
}

func TestPlan(t *testing.T) {
	i18n.SetLang("en")
	f := form()
	f.Enabled.On = false
	f.SSID.Value = "New"
	f.Channel.Set("11")
	f.Security.Set("none")
	sets := f.Plan()
	if len(sets) != 6 {
		t.Fatalf("sets: %+v", sets)
	}
	if sets[0] != (Set{"radio0", "disabled", "1"}) || sets[1] != (Set{"default_radio0", "disabled", "1"}) {
		t.Fatalf("disable: %+v", sets[:2])
	}
	// key must be cleared when the network becomes open
	last := sets[len(sets)-1]
	if last.Option != "key" || last.Value != "" {
		t.Fatalf("key not cleared: %+v", last)
	}

	clean := form()
	if len(clean.Plan()) != 0 {
		t.Fatal("clean form produced writes")
	}
}

func TestResetAdopt(t *testing.T) {
	f := form()
	f.SSID.Value = "temp"
	if !f.Changed() {
		t.Fatal("changed not detected")
	}
	f.Reset()
	if f.SSID.Value != "OpenWrt" || f.Changed() {
		t.Fatal("reset failed")
	}
	f.SSID.Value = "applied"
	f.Adopt()
	if f.Changed() {
		t.Fatal("adopt failed")
	}
}
