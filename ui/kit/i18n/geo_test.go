package i18n

import (
	"path/filepath"
	"testing"

	"openwrt-monitor-tui/task"
)

func TestGeoCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache")
	c := task.Open(path)
	if err := c.Put("geo", "jp"); err != nil {
		t.Fatal(err)
	}
	old := geoCache
	geoCache = func() *task.Cache { return c }
	t.Cleanup(func() { geoCache = old })

	// A cached answer is returned without touching the network.
	code, ok := geoCountry()
	if !ok || code != "JP" {
		t.Fatalf("geo: %q %v", code, ok)
	}
}

func TestParseCountryJSON(t *testing.T) {
	code, ok := parseCountryJSON(`{"ip":"1.2.3.4","city":"Tokyo","country":"JP"}`)
	if !ok || code != "JP" {
		t.Fatalf("geo json: %q %v", code, ok)
	}
	if _, ok := parseCountryJSON(`{"ip":"1.2.3.4"}`); ok {
		t.Fatal("missing country must not parse")
	}
}
