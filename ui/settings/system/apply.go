package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Run executes one action. Reboot and poweroff do not return on a healthy
// system; the error is shown only if the command refuses to run.
func Run(a Action) error {
	// Long enough for firstboot to finish wiping the overlay before the
	// reboot that follows it.
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	switch a {
	case Poweroff:
		return execCmd(ctx, "poweroff")
	case Factory:
		// firstboot resets the overlay; the reboot makes it take effect.
		if err := execCmd(ctx, "firstboot", "-y"); err != nil {
			return err
		}
		return execCmd(ctx, "reboot")
	default:
		return execCmd(ctx, "reboot")
	}
}

// execCmd is swapped by tests; it runs one command.
var execCmd = run

func run(ctx context.Context, name string, args ...string) error {
	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(string(out))
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	if msg == "" {
		return fmt.Errorf("%s: %w", name, err)
	}
	return fmt.Errorf("%s: %s", name, msg)
}
