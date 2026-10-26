package en

import "openwrt-monitor-tui/ui/i18n/table"

// T is the English string table.
var T = table.Strings{
	Tabs: [4]string{"Status", "Network", "Processes", "Settings"},

	Loading:         "loading…",
	HintScroll:      "tab switch  wheel/j/k scroll  r refresh  q quit",
	HintProcs:       "M by memory  P by cpu  wheel/j/k scroll  r refresh  q quit",
	HintSettings:    "j/k field  ←/→ band  enter edit  esc back  q quit",
	HintSidebar:     "j/k select  enter open  ←/→ tab  esc quit",
	HintLang:        "j/k choose  ←/→ tab  enter apply  esc back  q quit",
	HintCron:        "j/k field  ←/→ tab·value  enter edit  esc back  q quit",
	HintStartup:     "j/k task  space on/off  enter edit  a add  d delete  esc back  q quit",
	HintCronList:    "j/k select  enter edit  n new  esc back  q quit",
	HintSystem:      "j/k select  enter confirm  esc back  q quit",
	Boot:            "startup",
	UptimeFmt:       "%dd %02dh %02dm",
	RefreshInterval: "refresh",
	Language:        "Language",
	Auto:            "Auto",
	Cron:            "Scheduled",
	Startup:         "Startup",
	System:          "System",

	Hostname: "Hostname", Model: "Model", Kernel: "Kernel", Uptime: "Uptime",
	Load: "Load", CPU: "CPU", Memory: "Memory", Swap: "Swap", Temp: "Temp",
	FreeOf:  "%s free of %s",
	ClockNA: "clock n/a",
	FreqMax: "%s max",

	Iface: "IFACE", State: "STATE", Address: "ADDRESS", RxRate: "RX/s", TxRate: "TX/s",
	Up: "up", Down: "down",
	NoIfaces:     "no interfaces found",
	TotalTraffic: "total traffic  down %s  up %s",
	Traffic:      "TRAFFIC",

	NetTabOverview: "Overview", NetTabSpeed: "Speed test", NetTabPorts: "Ports",
	HintSpeed: "enter start  esc back  q quit",
	HintPorts: "enter start scan  esc back  q quit",

	SpeedStart: "Start", SpeedRunning: "testing…", SpeedTool: "tool",
	SpeedServer: "server", SpeedDown: "down", SpeedUp: "up", SpeedErr: "speed test failed",

	PortStart: "Start scan", PortRunning: "scanning…", PortProbed: "probed",
	PortSummary: "found %d hosts, %d open ports", PortNone: "no open ports found",

	PID: "PID", CPUPct: "CPU%", RSS: "RSS", Command: "COMMAND",
	NoProcs:     "no processes found",
	ProcSummary: "%d processes, showing top %d",

	NoRadio: "no wireless radio configured",

	Enabled: "Enabled", SSID: "SSID", Password: "Password",
	Channel: "Channel", Security: "Security",
	Apply: "Apply", Cancel: "Cancel", Applying: "applying…",
	Applied: "applied", Reverted: "reverted", NoChanges: "no changes",
	ApplyFirst: "apply or cancel first",
	NoAP:       "this radio has no AP section",

	ValChannel:   "channel: use auto or 1-233",
	ValSSIDEmpty: "SSID must not be empty",
	ValSSIDLong:  "SSID is longer than 32 bytes",
	ValPassword:  "password must be 8-63 characters",
	Hint24:       "auto, 1-13",
	Hint5:        "auto, 36-165",
	Hint6:        "auto, 1-233 step 4",
	Hint60:       "auto, 1-6",
	HintOther:    "auto",
	Open:         "open",

	CronTabList: "JOBS",
	CronTabNew:  "NEW",
	CronTabEdit: "EDIT",
	CronNone:    "no jobs yet",
	CronKind:    "Schedule",
	CronEvery:   "Interval",
	CronAt:      "Time",
	CronExpr:    "Expression",
	CronWeekday: "Weekday",
	CronCommand: "Command",
	CronAdded:   "job added",
	CronUpdated: "job updated",
	CronNoCron:  "no crontab on this device",
	CronKinds:   [5]string{"every N min", "every N hour", "daily", "weekly", "custom"},
	CronDays:    [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},

	StartupNoTasks: "no tasks yet",
	StartupSaved:   "saved; runs on next boot",
	StartupMiss:    "cannot read %s",

	ValCronRange: "interval out of range",
	ValCronTime:  "time must be HH:MM",
	ValCronExpr:  "expression needs 5 fields",

	SysHint:         "dangerous actions, confirmed twice",
	SysReboot:       "Reboot",
	SysRebootDesc:   "restart the device",
	SysPoweroff:     "Power off",
	SysPoweroffDesc: "cut the power",
	SysFactory:      "Factory reset",
	SysFactoryDesc:  "wipe the overlay and reboot",
	SysConfirm:      "confirm %s?  enter to run, esc to cancel",
	SysRunning:      "running…",
	SysDone:         "command submitted",
	ValCommand:      "command must not be empty",
}
