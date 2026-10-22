package collect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCron(t *testing.T) {
	src := `# comment
SHELL=/bin/sh
*/5 * * * * /usr/bin/logger hello
0 3 * * * /etc/init.d/dropbear restart
@reboot /bin/true
bad line
59 23 * * 6 echo "weekend,   spaced"
`
	jobs := parseCron(src)
	if len(jobs) != 3 {
		t.Fatalf("jobs: %+v", jobs)
	}
	if !jobs[0].Enabled || !jobs[1].Enabled {
		t.Fatalf("enabled flags: %+v", jobs)
	}
	if jobs[0].Schedule != "*/5 * * * *" || jobs[0].Command != "/usr/bin/logger hello" {
		t.Fatalf("first: %+v", jobs[0])
	}
	if jobs[2].Schedule != "59 23 * * 6" || jobs[2].Command != `echo "weekend,   spaced"` {
		t.Fatalf("spacing: %+v", jobs[2])
	}
}

func TestReadCronFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "root")
	body := "# managed\n*/10 * * * * /bin/true\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	info := readCronFile(path)
	if !info.Available || info.Path != path {
		t.Fatalf("info: %+v", info)
	}
	if len(info.Jobs) != 1 || info.Jobs[0].Command != "/bin/true" {
		t.Fatalf("jobs: %+v", info.Jobs)
	}
}

func TestReadCronMissing(t *testing.T) {
	info := readCronFile(filepath.Join(t.TempDir(), "nope"))
	if info.Available {
		t.Fatalf("missing crontab reported available: %+v", info)
	}
}

func TestReadStartup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rc.local")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	info := readStartupFile(path)
	if !info.Available || len(info.Content) != 2 || info.Content[1] != "exit 0" {
		t.Fatalf("startup: %+v", info)
	}

	missing := readStartupFile(filepath.Join(t.TempDir(), "absent"))
	if missing.Available || missing.Err == "" {
		t.Fatalf("missing file: %+v", missing)
	}
}

func TestParseDisabledJob(t *testing.T) {
	jobs := parseCron("#OFF# */5 * * * * /bin/true\n*/10 * * * * /bin/false\n")
	if len(jobs) != 2 {
		t.Fatalf("jobs: %+v", jobs)
	}
	if jobs[0].Enabled || jobs[0].Command != "/bin/true" || jobs[0].Schedule != "*/5 * * * *" {
		t.Fatalf("disabled job: %+v", jobs[0])
	}
	if !jobs[1].Enabled {
		t.Fatalf("enabled job: %+v", jobs[1])
	}
	if jobs[0].Line != 0 || jobs[1].Line != 1 {
		t.Fatalf("line numbers: %+v", jobs)
	}
}
