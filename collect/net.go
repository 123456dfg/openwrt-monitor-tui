package collect

import (
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ifaceMeta struct {
	name  string
	up    bool
	addrs []string
}

func (c *Collector) readIfaces(now time.Time) []Iface {
	counters := map[string][2]uint64{}
	for _, line := range strings.Split(readFile("/proc/net/dev"), "\n") {
		i := strings.IndexByte(line, ':')
		if i < 0 {
			continue
		}
		name := strings.TrimSpace(line[:i])
		f := strings.Fields(line[i+1:])
		if len(f) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(f[0], 10, 64)
		tx, _ := strconv.ParseUint(f[8], 10, 64)
		counters[name] = [2]uint64{rx, tx}
	}

	dt := now.Sub(c.prevNetAt).Seconds()
	prev := c.prevNet
	if !c.seeded {
		dt = 0
		c.seeded = true
	}
	c.prevNet = counters
	c.prevNetAt = now

	// Names, flags and addresses need netlink dumps (one per interface for
	// Addrs), so they are refreshed on a slower cadence than the counters.
	if stale(c.ifacesAt, now, c.ifaceEvery) {
		c.ifaces = readIfaceMeta()
		c.ifacesAt = now
	}

	out := make([]Iface, 0, len(c.ifaces))
	for _, mt := range c.ifaces {
		in := Iface{Name: mt.name, Up: mt.up, Addrs: mt.addrs}
		if ctr, ok := counters[mt.name]; ok {
			in.RxBytes, in.TxBytes = ctr[0], ctr[1]
			if p, ok := prev[mt.name]; ok && dt > 0 {
				in.RxRate = rate(ctr[0], p[0], dt)
				in.TxRate = rate(ctr[1], p[1], dt)
			}
		}
		out = append(out, in)
	}
	return out
}

// readIfaceMeta collects what only netlink knows. Loopback is skipped; the
// list is sorted so the cached copy keeps a stable order between samples.
func readIfaceMeta() []ifaceMeta {
	ifs, _ := net.Interfaces()
	live := make([]net.Interface, 0, len(ifs))
	for _, ifi := range ifs {
		if ifi.Flags&net.FlagLoopback == 0 {
			live = append(live, ifi)
		}
	}

	// Each Addrs() call is its own netlink dump; run them together.
	addrs := make([][]string, len(live))
	var wg sync.WaitGroup
	for i := range live {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got, _ := live[i].Addrs()
			for _, a := range got {
				if ipnet, ok := a.(*net.IPNet); ok {
					addrs[i] = append(addrs[i], ipnet.IP.String())
				}
			}
		}(i)
	}
	wg.Wait()

	out := make([]ifaceMeta, len(live))
	for i, ifi := range live {
		out[i] = ifaceMeta{name: ifi.Name, up: ifi.Flags&net.FlagUp != 0, addrs: addrs[i]}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

func rate(cur, prev uint64, dt float64) float64 {
	if cur < prev {
		return 0
	}
	return float64(cur-prev) / dt
}
