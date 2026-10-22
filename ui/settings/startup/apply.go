package startup

import (
	"fmt"
	"os"
	"path/filepath"

	"openwrt-monitor-tui/collect"
)

// startupFile is a variable so tests can point at a fixture.
var startupFile = collect.StartupPath

// Save writes the rendered file to /etc/rc.local. The file is replaced through
// a temporary file in the same directory, so a failed write cannot leave a
// half-written boot script behind, and the executable bit is preserved.
func Save(body string) error {
	path := startupFile()
	mode := os.FileMode(0o755)
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".rc.local-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name) // no-op after a successful rename

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
