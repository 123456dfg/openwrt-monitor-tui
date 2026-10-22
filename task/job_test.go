package task

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestJobRunsOnce(t *testing.T) {
	var calls atomic.Int32
	j := NewJob(func() int {
		calls.Add(1)
		return int(calls.Load())
	})

	v, ready := j.Get()
	if ready || v != 0 {
		t.Fatalf("first Get: %d %v", v, ready)
	}
	if j.Ready() {
		t.Fatal("Ready before the run finished")
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		v, ready = j.Get()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("job never finished")
		}
		time.Sleep(time.Millisecond)
	}
	if v != 1 || calls.Load() != 1 {
		t.Fatalf("value=%d calls=%d", v, calls.Load())
	}

	// a second Get does not run it again
	if v, _ := j.Get(); v != 1 || calls.Load() != 1 {
		t.Fatalf("rerun: value=%d calls=%d", v, calls.Load())
	}
}

func TestJobReload(t *testing.T) {
	var calls atomic.Int32
	j := NewJob(func() int { return int(calls.Add(1)) })
	j.Get()

	if v := j.Reload(); v != 1 {
		t.Fatalf("reload: %d", v)
	}
	if v := j.Reload(); v != 2 {
		t.Fatalf("second reload: %d", v)
	}
	if v, ready := j.Get(); !ready || v != 2 {
		t.Fatalf("after reload: %d %v", v, ready)
	}
}
