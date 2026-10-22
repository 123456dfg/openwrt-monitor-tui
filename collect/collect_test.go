package collect

import (
	"testing"
	"time"
)

func TestIntervals(t *testing.T) {
	cases := []struct {
		mhz               float64
		proc, iface, freq time.Duration
	}{
		{880, time.Second, time.Second, time.Second},
		{0, time.Second, time.Second, time.Second},
		{999, time.Second, time.Second, time.Second},
		{1000, 3 * time.Second, 2 * time.Second, time.Minute},
		{3418, 3 * time.Second, 2 * time.Second, time.Minute},
	}
	for _, c := range cases {
		p, i, f := intervalsFor(c.mhz)
		if p != c.proc || i != c.iface || f != c.freq {
			t.Fatalf("%v MHz: %v/%v/%v", c.mhz, p, i, f)
		}
	}
}

func TestClampPct(t *testing.T) {
	if ClampPct(-1) != 0 || ClampPct(101) != 100 || ClampPct(42) != 42 {
		t.Fatal("clamp")
	}
}

func TestProcStatFields(t *testing.T) {
	// after the comm: state ppid pgrp session tty tpgid flags minflt cminflt
	// majflt cmajflt utime stime ... vsize rss
	line := []byte(" S 0 1 1 0 -1 4194560 123 0 0 0 50 60 0 0 20 0 1 0 100 12345 678 18446744073709551615")
	ut, st, rss, ok := procStatFields(line)
	if !ok || ut != 50 || st != 60 || rss != 678 {
		t.Fatalf("got %d %d %d %v", ut, st, rss, ok)
	}
}

func TestScopeIsolation(t *testing.T) {
	c := NewCollector()
	net := c.Collect(ScopeNet)
	if net.Mem.Total != 0 || net.Procs != nil || net.Wifi.Supported {
		t.Fatal("network scope touched other sources")
	}

	procs := c.Collect(ScopeProcs)
	if len(procs.Procs) == 0 || procs.ProcN == 0 {
		t.Fatal("no processes collected")
	}
	if len(procs.Ifaces) != len(net.Ifaces) {
		t.Fatal("cached interfaces were lost")
	}

	status := c.Collect(ScopeStatus)
	if status.Mem.Total == 0 {
		t.Fatal("memory missing")
	}
	if len(status.Procs) != len(procs.Procs) {
		t.Fatal("cached processes were lost")
	}
	if status.Uptime <= 0 {
		t.Fatal("uptime missing")
	}
}

func TestJobsRunInBackground(t *testing.T) {
	c := NewCollector()
	boot := c.Collect(ScopeBoot)
	if boot.Cron.Loading || boot.Startup.Loading {
		t.Fatalf("boot sample touched the job files: %+v", boot.Cron)
	}
	if c.jobs.Ready() {
		t.Fatal("boot sample started the background reader")
	}

	first := c.Collect(ScopeCron | ScopeStartup)
	if !first.Cron.Loading || !first.Startup.Loading {
		t.Fatalf("first settings sample should be loading: cron=%+v", first.Cron)
	}

	deadline := time.Now().Add(3 * time.Second)
	for {
		s := c.Collect(ScopeCron | ScopeStartup)
		if !s.Cron.Loading {
			if s.Cron.Path == "" || s.Startup.Path == "" {
				t.Fatalf("background result incomplete: %+v %+v", s.Cron, s.Startup)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background read never finished")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// a synchronous reload updates the cache without a loading state
	cron, startup := c.ReloadJobs()
	if cron.Loading || startup.Loading || cron.Path == "" {
		t.Fatalf("reload: %+v %+v", cron, startup)
	}
	if got := c.Collect(ScopeCron); got.Cron.Loading {
		t.Fatal("sample after reload is loading")
	}
}
