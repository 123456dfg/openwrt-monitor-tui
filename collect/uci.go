package collect

import "strings"

type uciSection struct {
	typ  string
	name string
	opts map[string]string
}

// parseUCI is a deliberately small reader for /etc/config files: config,
// option and the quoting that matters. list lines are irrelevant here.
func parseUCI(path string) []uciSection {
	var out []uciSection
	for _, line := range strings.Split(readFile(path), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		f := strings.Fields(line)
		switch f[0] {
		case "config":
			if len(f) < 2 {
				continue
			}
			s := uciSection{typ: f[1], opts: map[string]string{}}
			if len(f) > 2 {
				s.name = unquote(f[2])
			}
			out = append(out, s)
		case "option":
			if len(out) == 0 || len(f) < 3 {
				continue
			}
			rest := strings.TrimSpace(line[len(f[0]):])
			i := strings.IndexAny(rest, " \t")
			if i < 0 {
				continue
			}
			cur := &out[len(out)-1]
			cur.opts[rest[:i]] = unquote(rest[i+1:])
		}
	}
	return out
}

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && (v[0] == '\'' || v[0] == '"') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}
	return v
}
