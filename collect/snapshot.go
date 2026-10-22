package collect

import "time"

// Snapshot is one point-in-time reading of the router state.
type Snapshot struct {
	At       time.Time
	Hostname string
	Model    string
	Release  string
	Arch     string
	Kernel   string

	Uptime time.Duration
	Load   [3]float64

	CPUUsage   float64
	CPUFreq    float64 // current clock, MHz; 0 when unknown
	CPUFreqMax float64 // highest clock the platform advertises, MHz; 0 when unknown
	Mem        MemInfo
	Disks      []DiskInfo

	Ifaces []Iface
	Procs  []Proc
	ProcN  int

	Wifi    WifiInfo
	Cron    CronInfo
	Startup StartupInfo

	TempC float64 // < 0 when unknown

	Took time.Duration // how long the sample took
}

type MemInfo struct {
	Total, Free, Available, Buffers, Cached uint64
	SwapTotal, SwapFree                     uint64
}

type DiskInfo struct {
	Mount  string
	FSType string
	Total  uint64
	Free   uint64
}

type Iface struct {
	Name    string
	Up      bool
	Addrs   []string
	RxBytes uint64
	TxBytes uint64
	RxRate  float64
	TxRate  float64
}

type Proc struct {
	PID  int
	RSS  uint64
	Name string
	CPU  float64 // percent of one core since the previous process sample

	ticks uint64 // utime+stime, internal for the delta
}

// WifiInfo is what the Settings tab can edit. Supported is hardware presence,
// not whether the radios are currently enabled.
type WifiInfo struct {
	Supported bool
	Bands     []WifiBand
}

// WifiBand is one radio plus its first AP: the edit form writes back to exactly
// these two UCI sections, so a multi-SSID radio only exposes its first AP.
type WifiBand struct {
	Name     string // "2.4 GHz", "5 GHz", ...; empty when the config does not say
	Radio    string // wifi-device section name
	Iface    string // wifi-iface section name, "" when the radio has no AP
	Enabled  bool   // radio and AP both enabled
	SSID     string
	Key      string
	Security string // raw encryption option, e.g. "psk2"
	Channel  string // raw channel option; "auto" when unset
	Width    int    // channel width in MHz from htmode, 0 when unknown
}

func (s Snapshot) TotalRx() uint64 {
	var n uint64
	for _, i := range s.Ifaces {
		n += i.RxBytes
	}
	return n
}

func (s Snapshot) TotalTx() uint64 {
	var n uint64
	for _, i := range s.Ifaces {
		n += i.TxBytes
	}
	return n
}
