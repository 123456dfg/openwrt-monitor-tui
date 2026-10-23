package procs

import (
	"fmt"
	"sort"
	"testing"

	"openwrt-monitor-tui/collect"
)

func benchProcs(n int) []collect.Proc {
	ps := make([]collect.Proc, n)
	for i := range ps {
		ps[i] = collect.Proc{
			PID:  i + 1,
			Name: fmt.Sprintf("proc%03d", i%50),
			RSS:  uint64((i*7919)%65536) * 1024,
			CPU:  float64((i*104729)%300) / 10,
		}
	}
	return ps
}

// BenchmarkTop is the ranking cost paid once per snapshot (or sort-key
// change), not per frame.
func BenchmarkTop150(b *testing.B) {
	ps := benchProcs(150)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Top(ps, ByCPU, true, DisplayLimit)
	}
}

func BenchmarkTop300(b *testing.B) {
	ps := benchProcs(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Top(ps, ByCPU, true, DisplayLimit)
	}
}

// BenchmarkBody is the per-frame render cost: it must not grow with the
// process count, only with DisplayLimit.
func BenchmarkBody(b *testing.B) {
	rows := Top(benchProcs(300), ByCPU, true, DisplayLimit)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Body(rows, 300, ByCPU, true)
	}
}

// BenchmarkSortPerFrame is what the render path used to pay on every keypress
// and redraw: a full copy plus a reflect-based sort of the whole table.
func BenchmarkSortPerFrame300(b *testing.B) {
	ps := benchProcs(300)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sorted := make([]collect.Proc, len(ps))
		copy(sorted, ps)
		sort.Slice(sorted, func(x, y int) bool { return sorted[x].CPU > sorted[y].CPU })
	}
}
