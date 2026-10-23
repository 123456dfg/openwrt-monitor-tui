package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/collect"
	"openwrt-monitor-tui/ui/kit"
	"openwrt-monitor-tui/ui/network"
	"openwrt-monitor-tui/ui/network/speedtest"
	"openwrt-monitor-tui/ui/procs"
	"openwrt-monitor-tui/ui/settings"
	"openwrt-monitor-tui/ui/settings/cron"
	"openwrt-monitor-tui/ui/settings/startup"
	"openwrt-monitor-tui/ui/settings/wifi"
	"openwrt-monitor-tui/ui/status"
	"openwrt-monitor-tui/ui/widgets/components"
)

const (
	chrome      = 6 // header, rule, tabs, rule, rule, status
	tabRowY     = 2 // line index of the tab bar in View
	bandRowY    = 4 // line index of the band strip: four chrome rows above the body
	mouseScroll = 3
	tabSettings = 3
)

// tabRefreshDelay lets the frame for a tab switch reach the terminal before the
// sample starts. On a slow single-core router a /proc walk right after the
// keypress would otherwise compete with drawing.
const tabRefreshDelay = 150 * time.Millisecond

type (
	snapshotMsg struct {
		snap collect.Snapshot
		tick bool
	}
	tickMsg      time.Time
	refreshMsg   struct{ seq int }
	wifiApplyMsg struct {
		err error
	}
	cronApplyMsg struct {
		err     error
		updated bool
	}
	cronToggleMsg struct {
		line int
		err  error
	}
	systemApplyMsg struct {
		err error
	}
	jobsReloadMsg struct {
		cron    collect.CronInfo
		startup collect.StartupInfo
	}
	startupApplyMsg struct {
		err error
	}
	speedPreparedMsg struct {
		native any
		server string
		err    error
	}
	speedDownMsg struct {
		rate speedtest.Rate
		err  error
	}
	speedUpMsg struct {
		rate speedtest.Rate
		err  error
	}
)

type model struct {
	col  *collect.Collector
	snap collect.Snapshot

	tab       int
	subSel    int
	subOpen   settings.Section
	subActive bool
	band      int
	scroll    int

	netSel    int  // network sidebar cursor: overview or speed test
	netOpen   int  // network sub-tab whose pane is shown
	netActive bool // the pane (not the sidebar) has the focus
	speed     speedtest.State

	procBy   procs.Sort // process table ranking column
	procDesc bool       // heaviest first; the default
	procRows []collect.Proc

	langSel   int
	langFocus int

	traffic *network.Tracker
	hist    *status.Tracker

	form     wifi.Form
	formNote string // transient "applied"/"reverted" feedback
	formErr  string // last validation or uci/wifi error

	cron       cron.Form
	cronTab    int
	cronCursor int
	cronNote   string
	cronErr    string

	startup     startup.Form
	startupNote string
	startupErr  string

	sysSel     int
	sysConfirm bool
	sysBusy    bool
	sysNote    string
	sysErr     string
	showKeys   bool

	w, h       int
	refresh    time.Duration
	boot       time.Duration // process start to first sample
	refreshSeq int
}

func tabNames() []string { return kit.Tr.Tabs[:] }

// tabHits holds the [start,end] x range of each tab, matching View's layout.
// Widths are terminal cells, not runes, so the CJK labels stay clickable.
func (m model) tabHits() [][2]int {
	// View prefixes the bar with one space.
	return components.Tabs{Names: tabNames(), Cur: m.tab}.Hits(1)
}

// New builds the UI model and takes the first sample. started is the process
// start time; the difference is shown in the tab bar as the startup cost.
func New(refresh time.Duration, started time.Time) tea.Model {
	m := model{
		col:       collect.NewCollector(),
		refresh:   refresh,
		boot:      time.Since(started),
		langSel:   kit.LangChoice(),
		langFocus: kit.LangChoice(),
		traffic:   network.NewTracker(),
		hist:      status.NewTracker(),
		cron:      cron.NewForm(),
	}
	// The crontab and rc.local are parsed in the background on demand, so the
	// first frame only waits for the cheap sources.
	m.snap = m.col.Collect(collect.ScopeBoot)
	m.hist.Add(m.snap)
	m.resortProcs()
	m.subOpen = m.sections()[0]
	m.syncForm()
	return m
}

func (m model) Init() tea.Cmd {
	return m.tickCmd()
}

// setTab switches the main tab. Entering Settings always starts on the
// sidebar, so the sub-tab is opened explicitly with Enter.
func (m *model) setTab(i int) {
	m.tab = i
	m.scroll = 0
	if i == 1 {
		// Settings convention: the sidebar starts preselected on the first
		// sub-tab, not entered. A running speed test keeps running; only the
		// selection resets.
		m.netSel = network.TabOverview
		m.netOpen = network.TabOverview
		m.netActive = false
	}
	if i == tabSettings {
		// Every entry starts from a clean slate: default section entered,
		// default band and focus, unsaved edits and messages dropped.
		m.resetSettings()
	}
}

// moveNetSel walks the network sidebar and previews that sub-tab, mirroring
// the settings sidebar.
func (m *model) moveNetSel(d int) {
	m.netSel = wrapIndex(m.netSel, d, network.TabSpeed+1)
	m.netOpen = m.netSel
}

// sortProcs ranks the process table by the given column. Selecting a new
// column resets to descending; selecting the active one flips the direction.
// The ranked view is cached in procRows: sorting happens here and on each new
// snapshot, never inside View, which runs on every keypress.
func (m *model) sortProcs(by procs.Sort) {
	if m.procBy == by {
		m.procDesc = !m.procDesc
	} else {
		m.procBy = by
		m.procDesc = true
	}
	m.resortProcs()
}

func (m *model) resortProcs() {
	m.procRows = procs.Top(m.snap.Procs, m.procBy, m.procDesc, procs.DisplayLimit)
}

// collectCmd samples only the sources the active tab shows. tick marks the
// periodic chain: ad-hoc refreshes must not schedule another timer.
// reloadJobsCmd re-reads the crontab and rc.local right after a write, so the
// list reflects the change without waiting for the next tick.
func (m model) reloadJobsCmd() tea.Cmd {
	col := m.col
	return func() tea.Msg {
		cron, startup := col.ReloadJobs()
		return jobsReloadMsg{cron: cron, startup: startup}
	}
}

func (m model) collectCmd(tick bool) tea.Cmd {
	col := m.col
	scope := scopeFor(m.tab)
	return func() tea.Msg { return snapshotMsg{snap: col.Collect(scope), tick: tick} }
}

func scopeFor(tab int) collect.Scope {
	switch tab {
	case 1:
		return collect.ScopeNet
	case 2:
		return collect.ScopeProcs
	case tabSettings:
		return collect.ScopeSettings
	}
	return collect.ScopeStatus
}

func (m model) tickCmd() tea.Cmd {
	return tea.Tick(m.refresh, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// refreshSoon schedules a one-off sample for the now-active tab. Only the last
// request wins, so tab-hopping does not queue a burst of samples.
func (m *model) refreshSoon() tea.Cmd {
	m.refreshSeq++
	seq := m.refreshSeq
	return tea.Tick(tabRefreshDelay, func(time.Time) tea.Msg { return refreshMsg{seq: seq} })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		return m, m.collectCmd(true)

	case refreshMsg:
		if msg.seq != m.refreshSeq {
			return m, nil
		}
		return m, m.collectCmd(false)

	case snapshotMsg:
		m.snap = msg.snap
		if m.tab == 1 {
			m.traffic.Add(msg.snap.Ifaces)
		}
		if m.tab == 0 {
			m.hist.Add(msg.snap)
		}
		if m.tab == 2 {
			// Re-rank once per snapshot, outside the render path.
			m.resortProcs()
		}
		m.syncForm()
		m.scroll = m.clampScroll(m.scroll)
		if !msg.tick {
			return m, nil
		}
		return m, m.tickCmd()

	case systemApplyMsg:
		m.sysBusy = false
		m.sysConfirm = false
		if msg.err != nil {
			m.sysErr = msg.err.Error()
			m.sysNote = ""
			return m, nil
		}
		m.sysErr = ""
		m.sysNote = kit.Tr.SysDone
		return m, nil

	case speedPreparedMsg:
		if msg.err != nil {
			m.speed.Running = false
			m.speed.Err = msg.err.Error()
			m.speed.Note = ""
			return m, nil
		}
		m.speed.Native = msg.native
		m.speed.Server = msg.server
		m.speed.Tool = "speedtest-go"
		m.speed.Down, m.speed.Up = speedtest.Rate{}, speedtest.Rate{}
		m.speed.Live = &speedtest.Live{}
		m.speed.Phase, m.speed.Running = speedtest.PhaseDown, true
		m.speed.Note, m.speed.Err = "", ""
		return m, m.speedDownCmd()

	case speedDownMsg:
		m.speed.Down = msg.rate
		if msg.err != nil {
			m.speed.Phase, m.speed.Running = speedtest.PhaseDone, false
			m.speed.Err = msg.err.Error()
			return m, nil
		}
		// Download done: continue straight into the upload measurement.
		m.speed.Phase, m.speed.Running = speedtest.PhaseUp, true
		return m, m.speedUpCmd()

	case speedUpMsg:
		m.speed.Up = msg.rate
		m.speed.Phase, m.speed.Running = speedtest.PhaseDone, false
		if msg.err != nil {
			m.speed.Err = msg.err.Error()
			m.speed.Note = ""
			return m, nil
		}
		return m, nil

	case jobsReloadMsg:
		m.snap.Cron = msg.cron
		m.snap.Startup = msg.startup
		m.syncCron()
		m.syncStartup()
		return m, nil

	case cronToggleMsg:
		if msg.err != nil {
			// Undo the optimistic flip.
			for i := range m.snap.Cron.Jobs {
				if m.snap.Cron.Jobs[i].Line == msg.line {
					m.snap.Cron.Jobs[i].Enabled = !m.snap.Cron.Jobs[i].Enabled
				}
			}
			m.cronErr = msg.err.Error()
			m.cronNote = ""
			return m, nil
		}
		m.cronErr = ""
		// Re-read right away so the optimistic flip is confirmed (or undone)
		// without waiting for the next tick.
		return m, m.reloadJobsCmd()

	case startupApplyMsg:
		m.startup.Busy = false
		if msg.err != nil {
			m.startupErr = msg.err.Error()
			m.startupNote = ""
			return m, nil
		}
		m.startup.Adopt()
		m.startupErr = ""
		m.startupNote = kit.Tr.StartupSaved
		return m, m.reloadJobsCmd()

	case cronApplyMsg:
		m.cron.Busy = false
		if msg.err != nil {
			m.cronErr = msg.err.Error()
			m.cronNote = ""
			return m, nil
		}
		m.cron = cron.NewForm()
		m.cronTab = cron.TabList
		m.cronErr = ""
		if msg.updated {
			m.cronNote = kit.Tr.CronUpdated
		} else {
			m.cronNote = kit.Tr.CronAdded
		}
		return m, m.reloadJobsCmd()

	case wifiApplyMsg:
		m.form.Busy = false
		if msg.err != nil {
			m.formErr = msg.err.Error()
			m.formNote = ""
			return m, nil
		}
		m.form.Adopt()
		m.formErr = ""
		m.formNote = kit.Tr.Applied
		return m, m.collectCmd(false)

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)
	}
	return m, nil
}

func (m model) rows() int {
	return max(1, m.h-chrome)
}

func (m model) clampScroll(v int) int {
	total := len(m.body())
	if last := total - m.rows(); v > last {
		v = last
	}
	if v < 0 {
		v = 0
	}
	return v
}

func (m model) body() []string {
	switch m.tab {
	case 1:
		st := m.speed
		st.Active = m.netActive && m.netOpen == network.TabSpeed
		return network.Body(m.snap, m.traffic, m.w, m.netSel, m.netOpen, m.netActive, st)
	case 2:
		return procs.Body(m.procRows, m.snap.ProcN, m.procBy, m.procDesc)
	case tabSettings:
		return settings.Body(m.settingsState())
	}
	return status.Body(m.snap, m.hist)
}

// wrapIndex moves i by d inside [0,n) and stays at 0 for an empty list.
func wrapIndex(i, d, n int) int {
	if n <= 0 {
		return 0
	}
	return ((i+d)%n + n) % n
}
