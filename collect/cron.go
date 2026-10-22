package collect

import (
	"os"
	"strings"
)

// CronInfo is the crontab the settings tab reads and appends to.
type CronInfo struct {
	Available bool   // a crontab or its directory exists
	Loading   bool   // the background read has not finished yet
	Path      string // file that was read (or would be written)
	Jobs      []CronJob
	Err       string
}

// OffPrefix marks a job that is kept in the file but disabled. It is a comment
// as far as cron is concerned, so a disabled job simply does not run.
const OffPrefix = "#OFF# "

// CronJob is one crontab line: the five schedule fields plus the command.
// Line is the raw file line it came from, so the editor can rewrite exactly
// that entry without touching comments or other jobs.
type CronJob struct {
	Schedule string // "*/5 * * * *"
	Command  string
	Line     int  // raw file line
	Enabled  bool // false when the line carries OffPrefix
}

// cronPaths are tried in order. OpenWrt keeps root's crontab in
// /etc/crontabs/root; /etc/crontab exists on some generic BusyBox builds.
var cronPaths = []string{"/etc/crontabs/root", "/etc/crontab"}

// CronPath is the file a new job would be appended to.
func CronPath() string { return cronPathFor(cronPaths) }

func cronPathFor(paths []string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if len(paths) > 0 {
		return paths[0]
	}
	return ""
}

func readCron() CronInfo { return readCronFile(CronPath()) }

func readCronFile(path string) CronInfo {
	info := CronInfo{Path: path}
	_, fileErr := os.Stat(path)
	_, dirErr := os.Stat("/etc/crontabs")
	if fileErr != nil && dirErr != nil {
		return info // no crontab at all: the UI says so and offers the form
	}
	info.Available = true
	if fileErr != nil {
		return info // directory exists but root has no crontab yet
	}
	info.Jobs = parseCron(readFile(path))
	return info
}

// parseCron keeps ordinary job lines and drops comments, blanks, environment
// assignments and @-shortcuts (BusyBox crond does not run those).
func parseCron(src string) []CronJob {
	var out []CronJob
	for n, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		enabled := true
		if rest, off := strings.CutPrefix(line, OffPrefix); off {
			line, enabled = strings.TrimSpace(rest), false
		}
		if line == "" || line[0] == '#' || line[0] == '@' {
			continue
		}
		schedule, command, ok := splitCronLine(line)
		if !ok {
			continue // VAR=value or a malformed line
		}
		out = append(out, CronJob{Schedule: schedule, Command: command, Line: n, Enabled: enabled})
	}
	return out
}

// ParseCronLine splits one crontab line into its schedule and command, using
// the same rules as the sampler.
func ParseCronLine(line string) (schedule, command string, ok bool) {
	return splitCronLine(strings.TrimSpace(line))
}

// splitCronLine takes the five schedule fields and returns the command exactly
// as written. Collapsing whitespace the way strings.Fields does would rewrite
// quoted arguments.
func splitCronLine(line string) (schedule, command string, ok bool) {
	fields := make([]string, 0, 5)
	i := 0
	for len(fields) < 5 {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		if i >= len(line) {
			return "", "", false
		}
		start := i
		for i < len(line) && line[i] != ' ' && line[i] != '\t' {
			i++
		}
		fields = append(fields, line[start:i])
	}
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i >= len(line) {
		return "", "", false
	}
	return strings.Join(fields, " "), line[i:], true
}

// StartupInfo is the content of /etc/rc.local, which the settings tab edits.
type StartupInfo struct {
	Available bool
	Loading   bool // the background read has not finished yet
	Path      string
	Content   []string
	Err       string
}

const startupPath = "/etc/rc.local"

// StartupPath is the file the startup pane edits.
func StartupPath() string { return startupPath }

func readStartup() StartupInfo { return readStartupFile(startupPath) }

func readStartupFile(path string) StartupInfo {
	info := StartupInfo{Path: path}
	raw, err := os.ReadFile(path)
	if err != nil {
		info.Err = err.Error()
		return info
	}
	info.Available = true
	text := strings.TrimRight(string(raw), "\n")
	if text != "" {
		info.Content = strings.Split(text, "\n")
	}
	return info
}
