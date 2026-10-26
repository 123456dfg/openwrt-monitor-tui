package task

import (
	"errors"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// CommonPorts is the default probe list: services typically worth knowing
// about on a home LAN.
var CommonPorts = []int{22, 23, 53, 80, 139, 443, 445, 3389, 8080, 8443, 9100}

// probeTimeout bounds a single TCP connect attempt. LAN hosts that refuse
// answer within microseconds; unreachable ones burn the whole timeout.
const probeTimeout = 500 * time.Millisecond

// probeWorkers bounds how many probes run at once.
const probeWorkers = 128

// dialer is swapped by tests; it reports whether a TCP connection succeeds.
var dialer = func(addr string, timeout time.Duration) bool {
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

// HostResult is one discovered neighbour and the ports that answered.
type HostResult struct {
	IP    string
	MAC   string
	Ports []int // open ports, ascending
}

// PortScanResult is one finished LAN scan.
type PortScanResult struct {
	Hosts []HostResult
	Err   error
}

// ScanProgress carries the live probe counters between the scanning
// goroutines and the UI.
type ScanProgress struct {
	Probed atomic.Int64
	Total  atomic.Int64
}

func (p *ScanProgress) reset(total int64) {
	p.Probed.Store(0)
	p.Total.Store(total)
}

// PortScan is a background LAN port scan. Start runs a fresh scan on the
// calling goroutine (use it from a tea.Cmd); Progress exposes the live
// counters of the current run and Result publishes the finished one.
type PortScan struct {
	job      *Job[PortScanResult]
	progress *ScanProgress
}

func NewPortScan() *PortScan {
	p := &PortScan{progress: &ScanProgress{}}
	p.job = NewJob(func() PortScanResult {
		hosts, err := p.run()
		return PortScanResult{Hosts: hosts, Err: err}
	})
	return p
}

// Progress returns the live counters of the current (or last) run.
func (p *PortScan) Progress() *ScanProgress { return p.progress }

// Start runs a fresh scan synchronously; call it from a tea.Cmd goroutine.
// It supersedes any older run.
func (p *PortScan) Start() { p.job.Reload() }

// Result returns the finished scan without ever starting one; ok is false
// while no scan has completed yet.
func (p *PortScan) Result() (PortScanResult, bool) { return p.job.Peek() }

// run discovers the ARP neighbours and probes the common ports on each,
// updating progress as probes complete.
func (p *PortScan) run() ([]HostResult, error) {
	hosts := readARP(arpFile)
	p.progress.reset(int64(len(hosts) * len(CommonPorts)))

	sem := make(chan struct{}, probeWorkers)
	var wg sync.WaitGroup
	res := make([]HostResult, len(hosts))
	for hi, h := range hosts {
		wg.Add(1)
		go func(hi int, h HostResult) {
			defer wg.Done()
			res[hi] = HostResult{IP: h.IP, MAC: h.MAC, Ports: p.probeHost(h.IP, &wg, sem)}
		}(hi, h)
	}
	wg.Wait()

	sort.Slice(res, func(i, j int) bool { return lessIP(res[i].IP, res[j].IP) })
	return res, nil
}

// probeHost dials every common port on one host and returns the open ones.
// It registers its probes on wg (the caller's wait group) and acquires sem
// for each, so one shared budget throttles the whole scan.
func (p *PortScan) probeHost(ip string, wg *sync.WaitGroup, sem chan struct{}) []int {
	var mu sync.Mutex
	ports := make([]int, 0, 4)
	var inner sync.WaitGroup
	for _, port := range CommonPorts {
		wg.Add(1)
		inner.Add(1)
		sem <- struct{}{}
		go func(port int) {
			defer wg.Done()
			defer inner.Done()
			defer func() { <-sem }()
			if dialer(net.JoinHostPort(ip, strconv.Itoa(port)), probeTimeout) {
				mu.Lock()
				ports = append(ports, port)
				mu.Unlock()
			}
			p.progress.Probed.Add(1)
		}(port)
	}
	inner.Wait()
	sort.Ints(ports)
	return ports
}

// arpFile is the kernel's neighbour table; a variable so tests can inject a
// fixture.
var arpFile = "/proc/net/arp"

// readARP parses the ARP table into complete entries, one per IP: entries
// without a resolved MAC (flags without ATF_COM, or the all-zero address)
// are neighbours the kernel has not actually talked to.
func readARP(path string) []HostResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []HostResult
	seen := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n")[1:] {
		f := strings.Fields(line)
		if len(f) < 6 || net.ParseIP(f[0]) == nil {
			continue
		}
		flags, err := strconv.ParseInt(f[2], 0, 32)
		if err != nil || flags&0x2 == 0 { // ATF_COM: entry is complete
			continue
		}
		mac := strings.ToLower(f[3])
		if mac == "00:00:00:00:00:00" {
			continue
		}
		if seen[f[0]] {
			continue
		}
		seen[f[0]] = true
		out = append(out, HostResult{IP: f[0], MAC: mac})
	}
	sort.Slice(out, func(i, j int) bool { return lessIP(out[i].IP, out[j].IP) })
	return out
}

// lessIP orders IPv4 addresses numerically, falling back to the string order
// for anything else.
func lessIP(a, b string) bool {
	ia, ib := net.ParseIP(a).To4(), net.ParseIP(b).To4()
	if ia == nil || ib == nil {
		return a < b
	}
	for i := range ia {
		if ia[i] != ib[i] {
			return ia[i] < ib[i]
		}
	}
	return false
}

// ErrNoNeighbours is returned when the ARP table has no usable entries.
var ErrNoNeighbours = errors.New("no lan neighbours found")
