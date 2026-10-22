package collect

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"openwrt-monitor-tui/task"
)

// Collector keeps the previous sample so it can compute rates.
type Collector struct {
	staticOnce sync.Once // hostname/model/release/kernel never change
	static     staticInfo

	// Each scope has its own lock. A Settings or Network sample must not queue
	// behind a long /proc walk: on a slow router that is exactly the delay the
	// user sees after a click. The snapshot itself is merged under mu.
	statusMu sync.Mutex
	prevCPU  cpuTimes
	freqDirs []string
	freqAt   time.Time

	netMu     sync.Mutex
	prevNet   map[string][2]uint64
	prevNetAt time.Time
	seeded    bool
	ifaces    []ifaceMeta
	ifacesAt  time.Time

	procsMu sync.Mutex
	procs   []Proc
	procN   int
	procsAt time.Time
	procCPU map[int]uint64

	// Refresh cadences. A sub-GHz SoC gets one second for every source;
	// faster machines keep the slower cadences from the hot path.
	procEvery  time.Duration
	ifaceEvery time.Duration
	freqEvery  time.Duration

	// The crontab and rc.local are parsed in a background task so the initial
	// sample and tab switches never wait for the filesystem.
	jobs *task.Job[jobsResult]

	mu   sync.Mutex // guards last
	last Snapshot   // previous sample, copied for scopes that are not refreshed
}

type staticInfo struct {
	hostname, model, release, arch, kernel string
}

type cpuTimes struct {
	idle, total uint64
}

func NewCollector() *Collector {
	proc, iface, freq := intervalsFor(cpuinfoMHz())
	return &Collector{
		prevNet:    map[string][2]uint64{},
		procEvery:  proc,
		ifaceEvery: iface,
		freqEvery:  freq,
		last:       Snapshot{TempC: -1},
		jobs: task.NewJob(func() jobsResult {
			return jobsResult{cron: readCronCached(), startup: readStartupCached()}
		}),
	}
}

// Scope selects the sources a sample touches. The UI asks only for the tab it
// is showing; the other fields keep their previous value in the snapshot, so
// rates and caches stay meaningful.
type Scope uint8

const (
	ScopeStatus Scope = 1 << iota
	ScopeNet
	ScopeProcs
	ScopeWifi
	ScopeCron
	ScopeStartup

	// ScopeSettings is everything the Settings tab can show.
	ScopeSettings = ScopeWifi | ScopeCron | ScopeStartup
	ScopeAll      = ScopeStatus | ScopeNet | ScopeProcs | ScopeSettings

	// ScopeBoot is the first sample: everything except the file-backed
	// settings pages, whose parsing runs in the background.
	ScopeBoot = ScopeStatus | ScopeNet | ScopeProcs | ScopeWifi
)

// intervalsFor picks the refresh cadence for the expensive sources. The clock
// comes from /proc/cpuinfo. Below 1 GHz every source is refreshed every second;
// the same applies when the file has no "cpu MHz" line (most MIPS routers), so
// the conservative cadence is the default and only a known-fast CPU gets the
// lazy one.
func intervalsFor(mhz float64) (proc, iface, freq time.Duration) {
	if mhz >= 1000 {
		return 3 * time.Second, 2 * time.Second, time.Minute
	}
	return time.Second, time.Second, time.Second
}

// Collect samples the requested scope. Sources are independent syscalls, so
// the ones in scope fan out; the rest are copied from the previous snapshot and
// are only used by tabs that are not visible.
//
// Scopes use separate locks so a cheap Wi-Fi read can complete while a process
// walk is still running on another core, and the merge at the end is the only
// shared critical section.
func (c *Collector) Collect(scope Scope) Snapshot {
	start := time.Now()
	c.staticOnce.Do(c.loadStatic)

	now := time.Now()

	// Uptime is one cheap read and is shown on the status page, but a stale
	// value there is what users notice first when tabs have been switched.
	uptime := readUptime()

	// Hold the scope locks from the first read through the merge, so a second
	// sample of the same scope cannot publish older values afterwards. Locks
	// are taken in a fixed order and are independent between scopes.
	haveStatus := scope&ScopeStatus != 0
	haveNet := scope&ScopeNet != 0
	haveProcs := scope&ScopeProcs != 0
	if haveStatus {
		c.statusMu.Lock()
	}
	if haveNet {
		c.netMu.Lock()
	}
	if haveProcs {
		c.procsMu.Lock()
	}

	c.mu.Lock()
	s := c.last
	c.mu.Unlock()

	var wg sync.WaitGroup
	if haveStatus {
		wg.Add(5)
		go func() { defer wg.Done(); s.Load = readLoad() }()
		go func() { defer wg.Done(); s.Mem = readMem() }()
		go func() {
			defer wg.Done()
			s.CPUUsage = c.readCPU()
			s.CPUFreq, s.CPUFreqMax = c.readCPUFreq(now)
		}()
		go func() { defer wg.Done(); s.Disks = readDisks() }()
		go func() { defer wg.Done(); s.TempC = readTemp() }()
	}
	if haveNet {
		wg.Add(1)
		go func() { defer wg.Done(); s.Ifaces = c.readIfaces(now) }()
	}
	if haveProcs {
		wg.Add(1)
		go func() { defer wg.Done(); s.Procs, s.ProcN = c.procList(now) }()
	}
	if scope&ScopeWifi != 0 {
		wg.Add(1)
		go func() { defer wg.Done(); s.Wifi = readWifi() }()
	}
	if scope&(ScopeCron|ScopeStartup) != 0 {
		res, ready := c.jobs.Get()
		if scope&ScopeCron != 0 {
			s.Cron = res.cron
			s.Cron.Loading = !ready
		}
		if scope&ScopeStartup != 0 {
			s.Startup = res.startup
			s.Startup.Loading = !ready
		}
	}
	wg.Wait()

	c.mu.Lock()
	c.last.Uptime = uptime
	if scope&ScopeStatus != 0 {
		c.last.Load = s.Load
		c.last.Mem = s.Mem
		c.last.CPUUsage, c.last.CPUFreq, c.last.CPUFreqMax = s.CPUUsage, s.CPUFreq, s.CPUFreqMax
		c.last.Disks = s.Disks
		c.last.TempC = s.TempC
	}
	if scope&ScopeNet != 0 {
		c.last.Ifaces = s.Ifaces
	}
	if scope&ScopeProcs != 0 {
		c.last.Procs, c.last.ProcN = s.Procs, s.ProcN
	}
	if scope&ScopeWifi != 0 {
		c.last.Wifi = s.Wifi
	}
	if scope&ScopeCron != 0 {
		c.last.Cron = s.Cron
	}
	if scope&ScopeStartup != 0 {
		c.last.Startup = s.Startup
	}
	if now.After(c.last.At) {
		c.last.At = now
	}
	c.last.Hostname = c.static.hostname
	c.last.Model = c.static.model
	c.last.Release = c.static.release
	c.last.Arch = c.static.arch
	c.last.Kernel = c.static.kernel
	out := c.last
	c.mu.Unlock()

	if haveProcs {
		c.procsMu.Unlock()
	}
	if haveNet {
		c.netMu.Unlock()
	}
	if haveStatus {
		c.statusMu.Unlock()
	}

	out.Took = time.Since(start)
	return out
}

// ReloadJobs reads the crontab and rc.local right away and refreshes the
// cache. It is meant to be called from a tea.Cmd goroutine after a write, so
// the UI never has to show a stale file for a whole tick.
func (c *Collector) ReloadJobs() (CronInfo, StartupInfo) {
	res := c.jobs.Reload()
	return res.cron, res.startup
}

// LowEnd reports whether /proc/cpuinfo has a clock below 1 GHz, or none at all.
// Unknown counts as low-end: most MIPS routers omit the "cpu MHz" line.
func LowEnd() bool { return cpuinfoMHz() < 1000 }

func (c *Collector) loadStatic() {
	c.static.hostname, _ = os.Hostname()
	c.static.model = strings.TrimSpace(readFile("/tmp/sysinfo/model"))
	rel := parseRelease("/etc/openwrt_release")
	c.static.release = rel["DISTRIB_DESCRIPTION"]
	c.static.arch = rel["DISTRIB_ARCH"]
	c.static.kernel = kernelRelease()
}

func readDisks() []DiskInfo {
	var out []DiskInfo
	seen := map[string]bool{}
	// "/" first, then the writable overlay if it is a separate mount.
	for _, m := range []string{"/", "/overlay"} {
		if seen[m] {
			continue
		}
		seen[m] = true
		var st syscall.Statfs_t
		if err := syscall.Statfs(m, &st); err != nil {
			continue
		}
		out = append(out, DiskInfo{
			Mount:  m,
			FSType: fsName(uint64(st.Type)),
			Total:  st.Blocks * uint64(st.Bsize),
			Free:   st.Bavail * uint64(st.Bsize),
		})
	}
	return out
}

func fsName(t uint64) string {
	switch t {
	case 0xEF53:
		return "ext4"
	case 0x9123683E:
		return "btrfs"
	case 0x58465342:
		return "xfs"
	case 0x73717368:
		return "squashfs"
	case 0x01021994:
		return "tmpfs"
	case 0x858458F6:
		return "ramfs"
	case 0x794C7630:
		return "overlay"
	case 0x65735546:
		return "fuse"
	}
	return ""
}

func readMem() MemInfo {
	var m MemInfo
	for _, line := range strings.Split(readFile("/proc/meminfo"), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		v, err := strconv.ParseUint(f[1], 10, 64)
		if err != nil {
			continue
		}
		v *= 1024
		switch f[0] {
		case "MemTotal:":
			m.Total = v
		case "MemFree:":
			m.Free = v
		case "MemAvailable:":
			m.Available = v
		case "Buffers:":
			m.Buffers = v
		case "Cached:":
			m.Cached = v
		case "SwapTotal:":
			m.SwapTotal = v
		case "SwapFree:":
			m.SwapFree = v
		}
	}
	if m.Available == 0 && m.Total > 0 {
		m.Available = m.Free + m.Buffers + m.Cached
	}
	return m
}

func readUptime() time.Duration {
	f := strings.Fields(readFile("/proc/uptime"))
	if len(f) == 0 {
		return 0
	}
	sec, err := strconv.ParseFloat(f[0], 64)
	if err != nil {
		return 0
	}
	return time.Duration(sec) * time.Second
}

func readLoad() [3]float64 {
	var l [3]float64
	f := strings.Fields(readFile("/proc/loadavg"))
	for i := 0; i < 3 && i < len(f); i++ {
		l[i], _ = strconv.ParseFloat(f[i], 64)
	}
	return l
}

func readTemp() float64 {
	raw := strings.TrimSpace(readFile("/sys/class/thermal/thermal_zone0/temp"))
	if raw == "" {
		return -1
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return -1
	}
	if v > 1000 {
		v /= 1000
	}
	return v
}

func parseRelease(path string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(readFile(path), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.Trim(strings.TrimSpace(v), `"'`)
	}
	return out
}

func kernelRelease() string {
	var u syscall.Utsname
	if err := syscall.Uname(&u); err != nil {
		return ""
	}
	b := make([]byte, 0, len(u.Release))
	for _, c := range u.Release {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}
	return string(b)
}

func readFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// stale reports whether a cached value is due for a refresh.
func stale(at, now time.Time, every time.Duration) bool {
	return at.IsZero() || now.Sub(at) >= every
}

func ClampPct(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
