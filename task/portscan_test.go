package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const arpSample = `IP address       HW type     Flags       HW address            Mask     Device
192.168.1.100    0x1         0x2         aa:bb:cc:dd:ee:01     *        br-lan
192.168.1.101    0x1         0x0         00:00:00:00:00:00     *        br-lan
192.168.1.20     0x1         0x2         AA:BB:CC:DD:EE:02     *        br-lan
192.168.1.100    0x1         0x2         aa:bb:cc:dd:ee:01     *        br-lan
junk line without fields
192.168.1.102    0x1         0x2         00:00:00:00:00:00     *        br-lan
`

func TestReadARP(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arp")
	if err := os.WriteFile(path, []byte(arpSample), 0o600); err != nil {
		t.Fatal(err)
	}
	hosts := readARP(path)
	// Incomplete (0x0), all-zero MAC and duplicate-IP entries are dropped;
	// what is left is ordered by IP.
	if len(hosts) != 2 {
		t.Fatalf("hosts: %+v", hosts)
	}
	if hosts[0].IP != "192.168.1.20" || hosts[0].MAC != "aa:bb:cc:dd:ee:02" {
		t.Fatalf("first host: %+v", hosts[0])
	}
	if hosts[1].IP != "192.168.1.100" {
		t.Fatalf("second host: %+v", hosts[1])
	}
}

func TestReadARPMissing(t *testing.T) {
	if hosts := readARP(filepath.Join(t.TempDir(), "nope")); len(hosts) != 0 {
		t.Fatalf("missing file: %+v", hosts)
	}
}

// With a fake dialer the scan must probe every host/port combination, report
// progress up to the total, and return hosts ordered by IP with only the
// open ports listed.
func TestStartScan(t *testing.T) {
	old := dialer
	t.Cleanup(func() { dialer = old })
	dialer = func(addr string, timeout time.Duration) bool {
		return strings.HasPrefix(addr, "192.168.1.20:") && strings.Contains(addr, ":22")
	}
	path := filepath.Join(t.TempDir(), "arp")
	if err := os.WriteFile(path, []byte(arpSample), 0o600); err != nil {
		t.Fatal(err)
	}
	oldARP := arpFile
	arpFile = path
	t.Cleanup(func() { arpFile = oldARP })

	p := NewPortScan()
	p.Start()
	res, ok := p.Result()
	if !ok {
		t.Fatal("scan did not publish a result")
	}
	if res.Err != nil {
		t.Fatalf("scan error: %v", res.Err)
	}
	if len(res.Hosts) != 2 {
		t.Fatalf("hosts: %+v", res.Hosts)
	}
	// 2 hosts x len(CommonPorts) probes must all be accounted for.
	if got := p.Progress().Probed.Load(); got != int64(2*len(CommonPorts)) {
		t.Fatalf("probed %d, want %d", got, 2*len(CommonPorts))
	}
	byIP := map[string]HostResult{}
	for _, h := range res.Hosts {
		byIP[h.IP] = h
	}
	if got := byIP["192.168.1.20"].Ports; len(got) != 1 || got[0] != 22 {
		t.Fatalf("192.168.1.20 ports: %v", got)
	}
	if got := byIP["192.168.1.100"].Ports; len(got) != 0 {
		t.Fatalf("192.168.1.100 ports: %v", got)
	}
}
