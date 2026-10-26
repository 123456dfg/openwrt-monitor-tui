// Package task runs work off the UI goroutine and keeps a small cache in /tmp.
package task

import "sync"

// Job produces a value in the background at most once, plus explicit reloads
// after a write. Get never blocks: it returns the last value and whether a run
// has finished.
type Job[T any] struct {
	mu      sync.Mutex
	produce func() T
	running bool
	have    bool
	gen     int
	value   T
}

func NewJob[T any](produce func() T) *Job[T] {
	return &Job[T]{produce: produce}
}

// Get starts the first run if needed and reports whether a result is ready.
func (j *Job[T]) Get() (T, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.running && !j.have {
		j.running = true
		gen := j.gen
		go func() {
			v := j.produce()
			j.mu.Lock()
			// A Reload may have started a newer run while this one was busy;
			// keep the newer result in that case.
			if gen == j.gen {
				j.value, j.have = v, true
			}
			j.running = false
			j.mu.Unlock()
		}()
	}
	return j.value, j.have
}

// Ready reports whether a result has been produced, without starting a run.
func (j *Job[T]) Ready() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.have
}

// Peek returns the last published value without starting a run. Unlike Get it
// never triggers the producer — for jobs that must start only on an explicit
// user action.
func (j *Job[T]) Peek() (T, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.value, j.have
}

// Reload runs the producer now and publishes the result. Call it from a
// tea.Cmd goroutine (not the UI loop); it supersedes any older background run.
func (j *Job[T]) Reload() T {
	j.mu.Lock()
	j.gen++
	j.running = true
	j.mu.Unlock()

	v := j.produce()

	j.mu.Lock()
	j.value, j.have = v, true
	j.running = false
	j.mu.Unlock()
	return v
}
