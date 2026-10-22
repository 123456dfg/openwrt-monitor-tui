package wifi

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Set is one option write; Section is a section name from
// /etc/config/wireless, e.g. "radio0" or "default_radio0".
type Set struct {
	Section string
	Option  string
	Value   string
}

// Apply writes the options and restarts the wireless stack, the same pair of
// steps LuCI performs. Values go as argv and never through a shell, so an SSID
// with spaces or quotes cannot turn into a command.
func Apply(sets []Set) error {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	for _, s := range sets {
		arg := fmt.Sprintf("wireless.%s.%s=%s", s.Section, s.Option, s.Value)
		if err := run(ctx, "uci", "set", arg); err != nil {
			return err
		}
	}
	if len(sets) > 0 {
		if err := run(ctx, "uci", "commit", "wireless"); err != nil {
			return err
		}
	}
	return run(ctx, "wifi", "reload")
}

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
		return fmt.Errorf("%s %s: %w", name, args[0], err)
	}
	return fmt.Errorf("%s %s: %s", name, args[0], msg)
}
