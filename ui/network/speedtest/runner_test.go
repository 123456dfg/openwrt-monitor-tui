package speedtest

import (
	"errors"
	"testing"
)

// fakeNative drives the NativeTest interface without touching the network.
type fakeNative struct {
	live   *Live
	handle any
}

func (f *fakeNative) Prepare() (any, string, error) {
	f.handle = "server"
	return f.handle, "Test City, TC", nil
}

func (f *fakeNative) Down(handle any, live *Live) (Rate, error) {
	f.live = live
	if handle != f.handle {
		return Rate{}, errors.New("bad handle")
	}
	live.set(false, 1000)
	return Rate{BPS: 1000, Bytes: 10000}, nil
}

func (f *fakeNative) Up(handle any, live *Live) (Rate, error) {
	if handle != f.handle {
		return Rate{}, errors.New("bad handle")
	}
	live.set(true, 500)
	return Rate{BPS: 500, Bytes: 5000}, nil
}

// The single provider is the native Ookla implementation and measures both
// directions.
func TestSingleProvider(t *testing.T) {
	if len(Providers) != 1 {
		t.Fatalf("providers: %d", len(Providers))
	}
	p := Providers[0]
	if p.Key != "net" || p.Name != "speedtest.net" || p.Native == nil {
		t.Fatalf("provider: %+v", p)
	}
}

func TestNativeFlow(t *testing.T) {
	fn := &fakeNative{}
	p := Provider{Key: "fake", Name: "fake", Native: fn}

	h, info, err := p.Native.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if info != "Test City, TC" {
		t.Fatalf("server info: %q", info)
	}
	var l Live
	down, err := p.Native.Down(h, &l)
	if err != nil || down.BPS != 1000 {
		t.Fatalf("down: %v %+v", err, down)
	}
	up, err := p.Native.Up(h, &l)
	if err != nil || up.BPS != 500 {
		t.Fatalf("up: %v %+v", err, up)
	}
	if fn.live != &l {
		t.Fatal("down did not receive the live handle")
	}
	if l.Down.Load() != 1000 || l.Up.Load() != 500 {
		t.Fatalf("live: %d %d", l.Down.Load(), l.Up.Load())
	}
}

func TestLiveSet(t *testing.T) {
	var l Live
	l.set(false, 1000)
	l.set(true, 500)
	if l.Down.Load() != 1000 || l.Up.Load() != 500 {
		t.Fatalf("live: down %d up %d", l.Down.Load(), l.Up.Load())
	}
	var nilLive *Live
	nilLive.set(false, 1) // must not panic
}
