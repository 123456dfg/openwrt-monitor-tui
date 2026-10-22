package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Cache is a tiny string map persisted as one quoted-key/quoted-value line per
// entry. It is small on purpose: the router build does not pull in encoding/json.
type Cache struct {
	path string

	mu     sync.Mutex
	values map[string]string
}

const cacheHeader = "openwrt-monitor-tui-cache 1"

// Open reads path if it exists; a missing or unreadable cache is simply empty.
func Open(path string) *Cache {
	c := &Cache{path: path, values: map[string]string{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		return c
	}
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != cacheHeader {
		return c
	}
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		space := strings.IndexByte(line, ' ')
		if space < 0 {
			continue
		}
		key, err1 := strconv.Unquote(line[:space])
		val, err2 := strconv.Unquote(line[space+1:])
		if err1 != nil || err2 != nil {
			continue
		}
		c.values[key] = val
	}
	return c
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.values[key]
	return v, ok
}

// Put stores a value and rewrites the file. Errors are reported but callers may
// treat the cache as best effort.
func (c *Cache) Put(key, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.values[key] == value {
		return nil
	}
	c.values[key] = value
	return c.saveLocked()
}

func (c *Cache) saveLocked() error {
	var b strings.Builder
	b.WriteString(cacheHeader)
	b.WriteByte('\n')
	for k, v := range c.values {
		fmt.Fprintf(&b, "%s %s\n", strconv.Quote(k), strconv.Quote(v))
	}

	dir := filepath.Dir(c.path)
	tmp, err := os.CreateTemp(dir, ".dpic-cache-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, 0o600); err != nil {
		return err
	}
	return os.Rename(name, c.path)
}

var (
	sharedOnce sync.Once
	shared     *Cache
	sharedPath = filepath.Join("/tmp", "openwrt-monitor-tui.cache")
)

// Shared is the process-wide cache at /tmp/openwrt-monitor-tui.cache.
func Shared() *Cache {
	sharedOnce.Do(func() { shared = Open(sharedPath) })
	return shared
}
