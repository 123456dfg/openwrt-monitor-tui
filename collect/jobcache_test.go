package collect

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"openwrt-monitor-tui/task"
)

func withCache(t *testing.T) *task.Cache {
	t.Helper()
	c := task.Open(filepath.Join(t.TempDir(), "cache"))
	old := jobCache
	jobCache = func() *task.Cache { return c }
	t.Cleanup(func() { jobCache = old })
	return c
}

func writeFile(t *testing.T, path, body string) time.Time {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return st.ModTime()
}

func TestCronCacheHit(t *testing.T) {
	cache := withCache(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "root")
	writeFile(t, path, "*/5 * * * * /bin/real\n")
	st, _ := os.Stat(path)

	// Pretend a previous run parsed a different job for this exact mtime.
	raw := encodeCron(CronInfo{Available: true, Path: path, Jobs: []CronJob{
		{Schedule: "1 2 * * *", Command: "/bin/cached", Line: 0, Enabled: true},
	}}, st.ModTime().UnixNano())
	if err := cache.Put("cron:"+path, raw); err != nil {
		t.Fatal(err)
	}

	info := readCronCachedFrom(path)
	if len(info.Jobs) != 1 || info.Jobs[0].Command != "/bin/cached" {
		t.Fatalf("cache not used: %+v", info)
	}
}

func TestCronCacheMissOnMtime(t *testing.T) {
	cache := withCache(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "root")
	writeFile(t, path, "*/5 * * * * /bin/real\n")
	st, _ := os.Stat(path)

	// A cache entry from an older mtime must be ignored and refreshed.
	raw := encodeCron(CronInfo{Available: true, Path: path, Jobs: []CronJob{
		{Schedule: "1 2 * * *", Command: "/bin/stale", Line: 0, Enabled: true},
	}}, st.ModTime().Add(-time.Minute).UnixNano())
	cache.Put("cron:"+path, raw)

	info := readCronCachedFrom(path)
	if len(info.Jobs) != 1 || info.Jobs[0].Command != "/bin/real" {
		t.Fatalf("stale cache used: %+v", info)
	}

	// The refresh is written back with the current mtime.
	raw, ok := cache.Get("cron:" + path)
	if !ok {
		t.Fatal("cache not refreshed")
	}
	got, mtime, ok := decodeCron(raw)
	if !ok || mtime != st.ModTime().UnixNano() || got.Jobs[0].Command != "/bin/real" {
		t.Fatalf("refreshed entry: %+v mtime=%d want=%d", got, mtime, st.ModTime().UnixNano())
	}
}

func TestCronCacheWriteIsReusable(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "cache")
	dir := t.TempDir()
	path := filepath.Join(dir, "root")
	writeFile(t, path, "# comment\n*/5 * * * * /bin/true\n\n0 3 * * * /bin/false\n")

	first := task.Open(cachePath)
	old := jobCache
	jobCache = func() *task.Cache { return first }
	info := readCronCachedFrom(path)
	jobCache = old
	if len(info.Jobs) != 2 {
		t.Fatalf("jobs: %+v", info)
	}

	// Reopen as a new process would.
	second := task.Open(cachePath)
	jobCache = func() *task.Cache { return second }
	t.Cleanup(func() { jobCache = old })
	again := readCronCachedFrom(path)
	if len(again.Jobs) != 2 || again.Jobs[1].Schedule != "0 3 * * *" ||
		again.Jobs[1].Command != "/bin/false" || again.Jobs[0].Line != 1 {
		t.Fatalf("reloaded cache: %+v", again)
	}
}

func TestStartupCache(t *testing.T) {
	cache := withCache(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "rc.local")
	writeFile(t, path, "#!/bin/sh\n/bin/real\nexit 0\n")
	st, _ := os.Stat(path)

	raw := encodeStartup(StartupInfo{Available: true, Path: path, Content: []string{"#!/bin/sh", "/bin/cached", "exit 0"}}, st.ModTime().UnixNano())
	cache.Put("startup:"+path, raw)
	info := readStartupCachedFrom(path)
	if len(info.Content) != 3 || info.Content[1] != "/bin/cached" {
		t.Fatalf("cache not used: %+v", info)
	}

	// A newer mtime forces a re-read and refreshes the entry.
	cache.Put("startup:"+path, encodeStartup(StartupInfo{Available: true, Path: path, Content: []string{"/bin/stale"}}, st.ModTime().Add(-time.Second).UnixNano()))
	info = readStartupCachedFrom(path)
	if len(info.Content) != 3 || info.Content[1] != "/bin/real" {
		t.Fatalf("stale cache used: %+v", info)
	}
}
