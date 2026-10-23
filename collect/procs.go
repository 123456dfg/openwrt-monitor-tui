package collect

import (
	"bytes"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// procList refreshes the process table on a slower cadence: opening every
// /proc/<pid>/stat is by far the most expensive part of a sample, and RSS
// rankings do not need to move every second. The full list is returned so the
// UI can re-sort it by CPU or memory without a resample.
func (c *Collector) procList(now time.Time) []Proc {
	if stale(c.procsAt, now, c.procEvery) {
		dt := 0.0
		if !c.procsAt.IsZero() {
			dt = now.Sub(c.procsAt).Seconds()
		}
		c.procs, c.procCPU = readProcs(c.procCPU, dt)
		c.procsAt = now
	}
	return c.procs
}

// procTicksPerSec is USER_HZ. Linux fixes it at 100 on every architecture,
// including the MIPS targets this runs on.
const procTicksPerSec = 100.0

// readProcs walks /proc with a small worker pool: every entry is a separate
// syscall, and on a router the wall time is dominated by that I/O, not CPU.
// The list comes back sorted by RSS; the UI re-sorts a copy as needed.
func readProcs(prev map[int]uint64, dt float64) ([]Proc, map[int]uint64) {
	dir, err := os.Open("/proc")
	if err != nil {
		return nil, nil
	}
	names, _ := dir.Readdirnames(-1)
	dir.Close()
	if len(names) == 0 {
		return nil, nil
	}

	idx := make([]int, 0, len(names)) // only PID dirs take part in the walk
	for i, name := range names {
		if name[0] >= '0' && name[0] <= '9' {
			idx = append(idx, i)
		}
	}
	if len(idx) == 0 {
		return nil, nil
	}

	found := make([]Proc, len(idx))
	workers := min(max(runtime.NumCPU()*2, 4), 16)
	workers = min(workers, len(idx))

	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			buf := make([]byte, 2048) // one stat line fits easily; reused per worker
			for i := range jobs {
				if p, ok := readProc(names[idx[i]], buf); ok {
					found[i] = p
				}
				// Hand the CPU back between batches: on a single-core router
				// this is what keeps keystrokes and redraws ahead of the walk.
				if i%8 == 7 {
					runtime.Gosched()
				}
			}
		}()
	}
	for i := range idx {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	next := make(map[int]uint64, len(idx))
	procs := make([]Proc, 0, len(idx))
	for _, p := range found {
		if p.Name == "" {
			continue
		}
		next[p.PID] = p.ticks
		if dt > 0 {
			if old, ok := prev[p.PID]; ok && p.ticks >= old {
				p.CPU = float64(p.ticks-old) / procTicksPerSec / dt * 100
			}
		}
		procs = append(procs, p)
	}
	sort.Slice(procs, func(i, j int) bool { return procs[i].RSS > procs[j].RSS })
	return procs, next
}

// readProc gets name and RSS from a single /proc/<pid>/stat read.
// Field order: pid (comm) state ppid ... ; comm may contain spaces and ')'.
// rss is field 24 of stat, i.e. index 21 after the closing paren.
//
// It uses a raw syscall and a caller-owned buffer: os.ReadFile allocates a
// fresh buffer, a File and a NUL-terminated path copy for every PID, which on
// a 100-process router is a steady stream of garbage.
func readProc(pid string, buf []byte) (Proc, bool) {
	num, err := strconv.Atoi(pid)
	if err != nil {
		return Proc{}, false
	}
	fd, err := syscall.Open("/proc/"+pid+"/stat", syscall.O_RDONLY|syscall.O_CLOEXEC, 0)
	if err != nil {
		return Proc{}, false
	}
	n, _ := syscall.Read(fd, buf)
	syscall.Close(fd)
	if n <= 0 {
		return Proc{}, false
	}
	s := buf[:n]

	l := bytes.IndexByte(s, '(')
	r := bytes.LastIndexByte(s, ')')
	if l < 0 || r < l {
		return Proc{}, false
	}
	name := string(s[l+1 : r])
	if name == "" {
		name = "?"
	}
	utime, stime, pages, ok := procStatFields(s[r+1:])
	if !ok {
		return Proc{}, false
	}
	return Proc{
		PID:   num,
		RSS:   pages * uint64(os.Getpagesize()),
		Name:  name,
		ticks: utime + stime,
	}, true
}

// procStatFields pulls utime (11), stime (12) and rss (21) out of the fields
// that follow the comm in /proc/<pid>/stat in a single pass.
func procStatFields(s []byte) (utime, stime, rss uint64, ok bool) {
	var seen [3]bool
	for i, field := 0, 0; ; field++ {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			return 0, 0, 0, false
		}
		start := i
		for i < len(s) && s[i] != ' ' && s[i] != '\t' && s[i] != '\n' {
			i++
		}
		switch field {
		case 11:
			utime, seen[0] = atou(s[start:i])
		case 12:
			stime, seen[1] = atou(s[start:i])
		case 21:
			rss, seen[2] = atou(s[start:i])
			if seen[0] && seen[1] && seen[2] {
				return utime, stime, rss, true
			}
			return 0, 0, 0, false
		}
	}
}

func atou(s []byte) (uint64, bool) {
	if len(s) == 0 {
		return 0, false
	}
	var v uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		v = v*10 + uint64(c-'0')
	}
	return v, true
}
