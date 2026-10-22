package collect

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (c *Collector) readCPU() float64 {
	f := strings.Fields(firstLine(readFile("/proc/stat")))
	if len(f) < 5 {
		return 0
	}
	var t cpuTimes
	for i, v := range f[1:] {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			break
		}
		t.total += n
		if i == 3 || i == 4 { // idle + iowait
			t.idle += n
		}
	}
	prev := c.prevCPU
	c.prevCPU = t
	if prev.total == 0 || t.total <= prev.total {
		return 0
	}
	busy := (t.total - prev.total) - (t.idle - prev.idle)
	return ClampPct(float64(busy) / float64(t.total-prev.total) * 100)
}

// readCPUFreq reports the CPU clock in MHz. cpufreq is the only real source on
// routers, and individual drivers expose it under different attribute names, so
// try all of them; x86 kernels instead print "cpu MHz" in /proc/cpuinfo. When
// none of those exist the kernel simply does not publish a clock (common on
// older MIPS targets): return 0 and let the UI say so.
func (c *Collector) readCPUFreq(now time.Time) (cur, maxFreq float64) {
	// The cpu* directory layout never changes at runtime, but scanning it with
	// filepath.Glob costs ~0.5ms, which is more than the actual sysfs reads.
	if stale(c.freqAt, now, c.freqEvery) {
		c.freqDirs, _ = filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq")
		c.freqAt = now
	}
	if mhz := cpufreqMHz(c.freqDirs, "scaling_cur_freq", "cpuinfo_cur_freq"); mhz > 0 {
		return mhz, 0
	}
	if mhz := cpuinfoMHz(); mhz > 0 {
		return mhz, 0
	}
	return 0, cpufreqMHz(c.freqDirs, "cpuinfo_max_freq", "scaling_max_freq")
}

// cpuinfoMHz scans for "cpu MHz" without strings.Split: /proc/cpuinfo is tens
// of KB on x86 and the split alone allocated more than the rest of the sample.
func cpuinfoMHz() float64 {
	best := 0.0
	s := readFile("/proc/cpuinfo")
	for len(s) > 0 {
		var line string
		if i := strings.IndexByte(s, '\n'); i >= 0 {
			line, s = s[:i], s[i+1:]
		} else {
			line, s = s, ""
		}
		// BogoMIPS is not a clock, so it never matches.
		k, v, ok := strings.Cut(line, ":")
		if !ok || strings.TrimSpace(k) != "cpu MHz" {
			continue
		}
		if mhz, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil && mhz > best {
			best = mhz
		}
	}
	return best
}

// cpufreqMHz returns the highest value among the named cpufreq attributes.
// Clocking is per-core, so the maximum is a better headline than an average.
func cpufreqMHz(dirs []string, names ...string) float64 {
	best := 0.0
	for _, dir := range dirs {
		for _, n := range names {
			khz, err := strconv.ParseFloat(strings.TrimSpace(readFile(filepath.Join(dir, n))), 64)
			if err != nil || khz <= 0 {
				continue
			}
			best = max(best, khz/1000)
			break
		}
	}
	return best
}
