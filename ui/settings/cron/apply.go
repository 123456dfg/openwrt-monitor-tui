package cron

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"openwrt-monitor-tui/collect"
)

// Both are variables so tests can use a fixture file and skip the init script.
var (
	cronPath   = collect.CronPath
	cronReload = reloadCron
)

// Add appends one job to the crontab and restarts cron. The line is written
// verbatim, so the caller must pass a validated schedule and command.
func Add(schedule, command string) error {
	path := cronPath()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, werr := f.WriteString(schedule + " " + command + "\n")
	cerr := f.Close()
	if werr != nil {
		return werr
	}
	if cerr != nil {
		return cerr
	}
	return cronReload()
}

// Update rewrites the line the job was read from, leaving every other line
// (comments included) untouched. origin is the job as loaded; if the line no
// longer holds it, the file changed underneath the editor and the update is
// refused instead of overwriting a different job.
func Update(line int, origin, schedule, command string) error {
	path := cronPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if line < 0 || line >= len(lines) {
		return fmt.Errorf("%s: job line is gone, refresh and retry", path)
	}
	if sched, cmd, ok := collect.ParseCronLine(lines[line]); !ok || sched+" "+cmd != origin {
		return fmt.Errorf("%s: job changed on disk, refresh and retry", path)
	}
	lines[line] = schedule + " " + command
	if err := writeAll(path, strings.Join(lines, "\n")+"\n"); err != nil {
		return err
	}
	return cronReload()
}

// SetEnabled comments or uncomments one job. The line keeps its position, so
// re-enabling restores it exactly where it was.
func SetEnabled(job collect.CronJob, on bool) error {
	path := cronPath()
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	if job.Line < 0 || job.Line >= len(lines) {
		return fmt.Errorf("%s: job line is gone, refresh and retry", path)
	}
	body := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[job.Line]), collect.OffPrefix))
	if sched, cmd, ok := collect.ParseCronLine(body); !ok || sched != job.Schedule || cmd != job.Command {
		return fmt.Errorf("%s: job changed on disk, refresh and retry", path)
	}
	line := job.Schedule + " " + job.Command
	if !on {
		line = collect.OffPrefix + line
	}
	lines[job.Line] = line
	if err := writeAll(path, strings.Join(lines, "\n")+"\n"); err != nil {
		return err
	}
	return cronReload()
}

// writeAll replaces the file through a temporary sibling, so a failed write
// cannot leave a half-written crontab behind. The mode is kept.
func writeAll(path, body string) error {
	mode := os.FileMode(0o600)
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".crontab-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)

	if _, err := tmp.WriteString(body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, mode); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// reloadCron asks the init script to pick up the new file. BusyBox crond re-reads
// the spool on SIGHUP, which is what the init script sends.
func reloadCron() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "/etc/init.d/cron", "restart").CombinedOutput()
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(string(out))
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	if msg == "" {
		return fmt.Errorf("/etc/init.d/cron restart: %w", err)
	}
	return fmt.Errorf("/etc/init.d/cron restart: %s", msg)
}
