package collect

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"openwrt-monitor-tui/task"
)

// jobCache is swapped by tests; production uses the shared /tmp file.
var jobCache = func() *task.Cache { return task.Shared() }

// jobsResult is what the background job produces.
type jobsResult struct {
	cron    CronInfo
	startup StartupInfo
}

// readCronCached reuses the parsed crontab while the file's mtime is unchanged.
func readCronCached() CronInfo { return readCronCachedFrom(CronPath()) }

func readCronCachedFrom(path string) CronInfo {
	st, err := os.Stat(path)
	if err != nil {
		return readCronFile(path) // missing or unreadable: nothing to cache
	}
	key := "cron:" + path
	if raw, ok := jobCache().Get(key); ok {
		if info, mtime, ok := decodeCron(raw); ok && mtime == st.ModTime().UnixNano() {
			return info
		}
	}
	info := readCronFile(path)
	jobCache().Put(key, encodeCron(info, st.ModTime().UnixNano()))
	return info
}

// readStartupCached reuses the parsed rc.local while its mtime is unchanged.
func readStartupCached() StartupInfo { return readStartupCachedFrom(startupPath) }

func readStartupCachedFrom(path string) StartupInfo {
	st, err := os.Stat(path)
	if err != nil {
		return readStartupFile(path)
	}
	key := "startup:" + path
	if raw, ok := jobCache().Get(key); ok {
		if info, mtime, ok := decodeStartup(raw); ok && mtime == st.ModTime().UnixNano() {
			return info
		}
	}
	info := readStartupFile(path)
	jobCache().Put(key, encodeStartup(info, st.ModTime().UnixNano()))
	return info
}

// The cache payload is line based: first the mtime, then the file path, then
// one quoted record per line. Quoting keeps commands with spaces and quotes
// unambiguous.
func encodeCron(info CronInfo, mtime int64) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d\n%q\n%d\n", mtime, info.Path, len(info.Jobs))
	for _, j := range info.Jobs {
		fmt.Fprintf(&b, "%q %q %d %t\n", j.Schedule, j.Command, j.Line, j.Enabled)
	}
	return b.String()
}

func decodeCron(raw string) (CronInfo, int64, bool) {
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	if len(lines) < 3 {
		return CronInfo{}, 0, false
	}
	mtime, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		return CronInfo{}, 0, false
	}
	var path string
	if _, err := fmt.Sscanf(lines[1], "%q", &path); err != nil {
		return CronInfo{}, 0, false
	}
	n, err := strconv.Atoi(lines[2])
	if err != nil || n < 0 || len(lines) != 3+n {
		return CronInfo{}, 0, false
	}
	jobs := make([]CronJob, 0, n)
	for _, line := range lines[3:] {
		var j CronJob
		if _, err := fmt.Sscanf(line, "%q %q %d %t", &j.Schedule, &j.Command, &j.Line, &j.Enabled); err != nil {
			return CronInfo{}, 0, false
		}
		jobs = append(jobs, j)
	}
	return CronInfo{Available: true, Path: path, Jobs: jobs}, mtime, true
}

func encodeStartup(info StartupInfo, mtime int64) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d\n%q\n%d\n", mtime, info.Path, len(info.Content))
	for _, line := range info.Content {
		fmt.Fprintf(&b, "%q\n", line)
	}
	return b.String()
}

func decodeStartup(raw string) (StartupInfo, int64, bool) {
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	if len(lines) < 3 {
		return StartupInfo{}, 0, false
	}
	mtime, err := strconv.ParseInt(lines[0], 10, 64)
	if err != nil {
		return StartupInfo{}, 0, false
	}
	var path string
	if _, err := fmt.Sscanf(lines[1], "%q", &path); err != nil {
		return StartupInfo{}, 0, false
	}
	n, err := strconv.Atoi(lines[2])
	if err != nil || n < 0 || len(lines) != 3+n {
		return StartupInfo{}, 0, false
	}
	content := make([]string, 0, n)
	for _, line := range lines[3:] {
		var text string
		if _, err := fmt.Sscanf(line, "%q", &text); err != nil {
			return StartupInfo{}, 0, false
		}
		content = append(content, text)
	}
	return StartupInfo{Available: true, Path: path, Content: content}, mtime, true
}
