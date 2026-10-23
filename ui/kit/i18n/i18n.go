// Package i18n owns every piece of user-visible text and the language
// selection behind it. It is a separate package so the string tables can be
// split per language without dragging the style helpers along; the kit package
// re-exports the entry points.
package i18n

import "strings"

// Strings is every piece of text the interface shows. Keeping it a struct
// instead of a map means a missing translation is a compile error, not a blank
// label at runtime. Values with verbs are used with fmt.Sprintf.
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

	NetTabOverview, NetTabSpeed string // network sidebar: overview / speed test
	HintSpeed                   string

	SpeedStart, SpeedRunning, SpeedTool string
	SpeedServer                         string
	SpeedDown, SpeedUp, SpeedErr        string

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

// Tr is the active translation. It is a pointer that kit shares: the setters
// below replace the pointed-to value rather than the pointer, so every reader
// sees the new language without any re-export step.
var Tr = new(Strings)

func init() { *Tr = en }

// langPick is what the user asked for (auto/en/zh), langAuto is what auto
// resolved to at startup. They are separate so switching to "auto" in the
// settings tab never has to hit the network again.
var (
	langPick     = "auto"
	langAutoPick = "en"
)

// SetLang is the startup entry point: "auto" asks an IP geolocation service
// which country the public address belongs to, falling back to the locale
// variables when the lookup fails or the router is offline.
func SetLang(name string) {
	langPick = normalizeLang(name)
	if langPick == "auto" {
		langAutoPick = autoLangCode()
	}
	apply()
}

// SetLangChoice switches language at runtime without any network access.
func SetLangChoice(name string) {
	langPick = normalizeLang(name)
	apply()
}

// apply swaps the active strings in place, through the shared pointer.
func apply() { *Tr = stringsFor(resolveLang()) }

// LangChoice is the index into langChoices of the current pick.
func LangChoice() int {
	for i, c := range langChoices {
		if c == langPick {
			return i
		}
	}
	return 0
}

// LangCode is the language actually in effect ("en" or "zh").
func LangCode() string { return resolveLang() }

// LangAutoCode is what "auto" resolved to at startup.
func LangAutoCode() string { return langAutoPick }

var langChoices = []string{"auto", "en", "zh"}

func resolveLang() string {
	if langPick == "auto" {
		return langAutoPick
	}
	return langPick
}

func stringsFor(code string) Strings {
	if code == "zh" {
		return zh
	}
	return en
}

func normalizeLang(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "zh", "zh-cn", "zh_cn", "cn", "chinese":
		return "zh"
	case "en", "english":
		return "en"
	}
	return "auto"
}
