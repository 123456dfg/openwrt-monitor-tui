package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache")
	c := Open(path)

	if _, ok := c.Get("missing"); ok {
		t.Fatal("missing key reported present")
	}
	tricky := "line one\nline two\twith tab \"quotes\" 中文"
	if err := c.Put("k", tricky); err != nil {
		t.Fatalf("put: %v", err)
	}
	if err := c.Put("geo", "CN"); err != nil {
		t.Fatalf("put geo: %v", err)
	}

	again := Open(path)
	if v, ok := again.Get("k"); !ok || v != tricky {
		t.Fatalf("round trip: %q %v", v, ok)
	}
	if v, ok := again.Get("geo"); !ok || v != "CN" {
		t.Fatalf("geo: %q %v", v, ok)
	}
}

func TestCacheMissingAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	if c := Open(filepath.Join(dir, "absent")); c == nil {
		t.Fatal("nil cache")
	}

	path := filepath.Join(dir, "bad")
	if err := os.WriteFile(path, []byte("not a cache\nwhatever\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	c := Open(path)
	if _, ok := c.Get("whatever"); ok {
		t.Fatal("corrupt header accepted")
	}
	// it still works as an in-memory cache and rewrites a valid file
	if err := c.Put("a", "b"); err != nil {
		t.Fatal(err)
	}
	if v, ok := Open(path).Get("a"); !ok || v != "b" {
		t.Fatalf("rewrite: %q %v", v, ok)
	}
}

func TestCacheSkipsUnchangedWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache")
	c := Open(path)
	if err := c.Put("a", "1"); err != nil {
		t.Fatal(err)
	}
	st1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Put("a", "1"); err != nil {
		t.Fatal(err)
	}
	st2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Fatal("unchanged value rewrote the file")
	}
}
