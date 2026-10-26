// Package table holds the Strings schema shared by every language table.
// It lives below the language packages so en/ and zh/ can define their tables
// without importing (and being imported by) the i18n core.
package table

type Strings struct {
	Tabs [4]string

	Loading         string
	HintScroll      string
	HintProcs       string
	HintSettings    string
	HintSidebar     string
	HintLang        string
	Boot            string // startup cost label in the tab bar
	UptimeFmt       string // fmt for dd/hh/mm
	RefreshInterval string

	Language string // settings sub-tab and language page title
	Auto     string // "auto" choice in the language picker
	Cron     string // settings sub-tab: scheduled jobs
	Startup  string // settings sub-tab: boot-time file
	System   string // settings sub-tab: reboot / power off / factory reset

	Hostname, Model, Kernel, Uptime, Load, CPU, Memory, Swap, Temp string
	FreeOf                                                         string
	ClockNA, FreqMax                                               string

	Iface, State, Address, RxRate, TxRate string
	Up, Down                              string
	NoIfaces                              string
	TotalTraffic                          string
	Traffic                               string

	NetTabOverview, NetTabSpeed, NetTabPorts string // network sidebar: overview / speed test / ports
	HintSpeed, HintPorts                     string

	SpeedStart, SpeedRunning, SpeedTool string
	SpeedServer                         string
	SpeedDown, SpeedUp, SpeedErr        string

	PortStart, PortRunning, PortProbed string
	PortSummary, PortNone              string

	PID, CPUPct, RSS, Command string
	NoProcs                   string
	ProcSummary               string

	NoRadio string

	Enabled, SSID, Password, Channel, Security string
	Apply, Cancel, Applying                    string
	Applied, Reverted, NoChanges, ApplyFirst   string
	NoAP                                       string

	ValChannel, ValSSIDEmpty, ValSSIDLong, ValPassword string
	Hint24, Hint5, Hint6, Hint60, HintOther            string
	Open                                               string

	HintCron     string
	HintCronList string
	HintStartup  string

	CronTabList    string // sub-tab: job list
	CronTabNew     string // sub-tab: new job
	CronTabEdit    string // sub-tab while an existing job is loaded
	CronNone       string // no job lines in the file
	CronKind       string // schedule-kind row
	CronEvery      string // "interval" row label
	CronAt         string // "time" row label
	CronExpr       string // "expression" row label for custom schedules
	CronWeekday    string
	CronCommand    string
	CronAdded      string
	CronUpdated    string
	CronNoCron     string
	CronKinds      [5]string // every-N-minutes, every-N-hours, daily, weekly, custom
	CronDays       [7]string
	StartupNoTasks string
	StartupSaved   string
	StartupMiss    string // fmt with the path
	ValCronRange   string
	ValCronTime    string
	ValCronExpr    string
	ValCommand     string

	HintSystem                      string
	SysHint                         string
	SysReboot, SysRebootDesc        string
	SysPoweroff, SysPoweroffDesc    string
	SysFactory, SysFactoryDesc      string
	SysConfirm, SysRunning, SysDone string
}
