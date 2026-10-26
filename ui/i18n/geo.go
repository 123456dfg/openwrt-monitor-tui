package i18n

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"openwrt-monitor-tui/task"
)

// autoLangCode prefers the public IP's country over the locale, because the
// locale on a router is frequently "C" even when the box sits in China.
func autoLangCode() string {
	if code, ok := geoCountry(); ok {
		switch code {
		case "CN", "TW", "HK", "MO":
			return "zh"
		}
		return "en"
	}
	if isZhLocale() {
		return "zh"
	}
	return "en"
}

// geoCache is swapped by tests; production shares the /tmp cache file.
var geoCache = func() *task.Cache { return task.Shared() }

// geoCountry answers from the cache when the lookup has been done before; the
// country rarely changes, so no freshness check is needed.
func geoCountry() (string, bool) {
	if code, ok := geoCache().Get("geo"); ok && len(code) == 2 && isAlpha(code) {
		return strings.ToUpper(code), true
	}
	code, ok := probeCountry()
	if !ok {
		return "", false
	}
	geoCache().Put("geo", code)
	return code, true
}

// probeCountry asks ipinfo.io which country the public address belongs to. The
// request goes through the system curl, falling back to wget when curl is
// missing or cannot complete (old build without TLS, blocked port, ...); a
// router without either tool simply falls back to the locale.
func probeCountry() (string, bool) {
	const url = "https://ipinfo.io/json"
	if path, err := exec.LookPath("curl"); err == nil {
		if out, ok := runGeo(path, "-fsSL", "--max-time", "3", url); ok {
			if code, ok := parseCountryJSON(out); ok {
				return code, true
			}
		}
	}
	if path, err := exec.LookPath("wget"); err == nil {
		if out, ok := runGeo(path, "-q", "-T", "3", "-O", "-", url); ok {
			if code, ok := parseCountryJSON(out); ok {
				return code, true
			}
		}
	}
	return "", false
}

func runGeo(path string, args ...string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, args...).Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

// parseCountryJSON digs the country field out of ipinfo's JSON without pulling
// in encoding/json for a three-field document.
func parseCountryJSON(s string) (string, bool) {
	i := strings.Index(s, `"country"`)
	if i < 0 {
		return "", false
	}
	rest := s[i+len(`"country"`):]
	j := strings.IndexByte(rest, ':')
	if j < 0 {
		return "", false
	}
	rest = strings.TrimSpace(rest[j+1:])
	if len(rest) < 3 || rest[0] != '"' {
		return "", false
	}
	end := strings.IndexByte(rest[1:], '"')
	if end < 0 {
		return "", false
	}
	code := rest[1 : 1+end]
	if len(code) != 2 || !isAlpha(code) {
		return "", false
	}
	return strings.ToUpper(code), true
}

func isAlpha(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}

func isZhLocale() bool {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		v := strings.ToLower(os.Getenv(k))
		if v == "" {
			continue
		}
		return strings.Contains(v, "zh")
	}
	return false
}
