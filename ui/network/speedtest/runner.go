// Package speedtest is the Speed test feature of the Network tab. It runs the
// Speedtest.net protocol in-process through speedtest-go — no external tools
// involved — measuring download and upload against the closest server, with
// live progress reported to the pane through atomic counters.
package speedtest

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	ookla "github.com/showwin/speedtest-go/speedtest"
)

// Measurement phases, mirrored by the pane.
const (
	PhaseIdle = iota
	PhaseDown
	PhaseUp
	PhaseDone
)

// Provider is one speed test source. Only Speedtest.net exists, backed by the
// speedtest-go library; the provider layer remains so the pane and the mouse
// layer keep working against an interface.
type Provider struct {
	Key    string // stable id
	Name   string // display name, e.g. "speedtest.net"
	Native NativeTest
}

// NativeTest is the library-backed test: Prepare selects a server once and
// describes it (city, country) for the pane, then Down and Up measure the two
// directions against it, feeding their running average into live for the
// real-time display.
type NativeTest interface {
	Prepare() (handle any, info string, err error)
	Down(handle any, live *Live) (Rate, error)
	Up(handle any, live *Live) (Rate, error)
}

// Providers lists the test sources. A single entry: Speedtest.net, which
// measures both directions.
var Providers = []Provider{
	{Key: "net", Name: "speedtest.net", Native: &ooklaTest{}},
}

// Rate is one measured direction: average bytes/s and the bytes moved.
type Rate struct {
	BPS   float64
	Bytes int64
}

// Live carries the running average of an in-flight measurement between the
// measuring goroutine and the UI. The pane polls it on every frame, so the
// numbers move in real time without extra messages.
type Live struct {
	Down, Up atomic.Int64 // smoothed bytes/s
}

func (l *Live) set(upload bool, v int64) {
	if l == nil {
		return
	}
	if upload {
		l.Up.Store(v)
	} else {
		l.Down.Store(v)
	}
}

// ooklaTest runs the Speedtest.net protocol through speedtest-go. The handle
// is the selected *ookla.Server; the client's rate callbacks drive the live
// display while a direction is in flight.
type ooklaTest struct {
	client *ookla.Speedtest
}

func (o *ooklaTest) Prepare() (any, string, error) {
	if o.client == nil {
		o.client = ookla.New()
	}
	serverList, err := o.client.FetchServers()
	if err != nil {
		return nil, "", err
	}
	targets, err := serverList.FindServer([]int{}) // empty = closest server
	if err != nil || len(targets) == 0 {
		return nil, "", errors.New("no test server")
	}
	s := targets[0]
	if err := s.PingTest(nil); err != nil {
		return nil, "", err
	}
	info := strings.TrimSpace(s.Name)
	if s.Country != "" && s.Country != s.Name {
		info += ", " + s.Country
	}
	return s, info, nil
}

func (o *ooklaTest) Down(handle any, live *Live) (Rate, error) {
	s := handle.(*ookla.Server)
	o.client.SetCallbackDownload(func(r ookla.ByteRate) { live.set(false, int64(r)) })
	start := time.Now()
	if err := s.DownloadTest(); err != nil {
		return Rate{}, err
	}
	bps := float64(s.DLSpeed)
	return Rate{BPS: bps, Bytes: int64(bps * time.Since(start).Seconds())}, nil
}

func (o *ooklaTest) Up(handle any, live *Live) (Rate, error) {
	s := handle.(*ookla.Server)
	o.client.SetCallbackUpload(func(r ookla.ByteRate) { live.set(true, int64(r)) })
	start := time.Now()
	if err := s.UploadTest(); err != nil {
		return Rate{}, err
	}
	bps := float64(s.ULSpeed)
	return Rate{BPS: bps, Bytes: int64(bps * time.Since(start).Seconds())}, nil
}
