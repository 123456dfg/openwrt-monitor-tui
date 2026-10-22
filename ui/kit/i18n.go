package kit

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"

	"openwrt-monitor-tui/task"
)

// Strings is every piece of text the interface shows. Keeping it a struct
// instead of a map means a missing translation is a compile error, not a blank
// label at runtime. Values with verbs are used with fmt.Sprintf.
type Strings struct {
	Tabs [4]string

	Loading         string
	HintScroll      string
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

var en = Strings{
	Tabs: [4]string{"Status", "Network", "Processes", "Settings"},

	Loading:         "loading…",
	HintScroll:      "tab switch  wheel/j/k scroll  r refresh  q quit",
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

	PID: "PID", CPUPct: "CPU%", RSS: "RSS", Command: "COMMAND",
	NoProcs:     "no processes found",
	ProcSummary: "%d processes, top %d by memory",

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

var zh = Strings{
	Tabs: [4]string{"状态", "网络", "进程", "设置"},

	Loading:         "加载中…",
	HintScroll:      "tab 切页  ↓/j 滚动  r 刷新  q 退出",
	HintSettings:    "j/k 字段  ←/→ 频段  enter 编辑  esc 返回  q 退出",
	HintSidebar:     "j/k 选择  enter 进入  ←/→ 切页  esc 退出",
	HintLang:        "j/k 选择  ←/→ 切页  enter 应用  esc 返回  q 退出",
	HintCron:        "j/k 字段  ←/→ 切换/取值  enter 编辑  esc 返回  q 退出",
	HintStartup:     "j/k 任务  space 启停  enter 编辑  a 新增  d 删除  esc 返回  q 退出",
	HintCronList:    "j/k 选择  enter 编辑  n 新建  esc 返回  q 退出",
	HintSystem:      "j/k 选择  enter 确认  esc 返回  q 退出",
	Boot:            "启动",
	UptimeFmt:       "%d天 %02d小时 %02d分钟",
	RefreshInterval: "刷新间隔",
	Language:        "语言",
	Auto:            "自动",
	Cron:            "定时任务",
	Startup:         "开机任务",
	System:          "系统",

	Hostname: "主机名", Model: "型号", Kernel: "内核", Uptime: "在线时间",
	Load: "负载", CPU: "CPU", Memory: "内存", Swap: "交换", Temp: "温度",
	FreeOf:  "%s 可用 / 共 %s",
	ClockNA: "频率未知",
	FreqMax: "最高 %s",

	Iface: "接口", State: "状态", Address: "地址", RxRate: "接收/s", TxRate: "发送/s",
	Up: "启用", Down: "停用",
	NoIfaces:     "未找到网络接口",
	TotalTraffic: "累计流量  下行 %s  上行 %s",
	Traffic:      "流量",

	PID: "进程号", CPUPct: "CPU%", RSS: "内存", Command: "命令",
	NoProcs:     "未找到进程",
	ProcSummary: "共 %d 个进程，按内存排序前 %d",

	NoRadio: "没有可用的无线 radio",

	Enabled: "启用", SSID: "SSID", Password: "密码",
	Channel: "信道", Security: "加密方式",
	Apply: "应用", Cancel: "取消", Applying: "应用中…",
	Applied: "已应用", Reverted: "已还原", NoChanges: "没有修改",
	ApplyFirst: "请先应用或取消",
	NoAP:       "该 radio 没有 AP 配置段",

	ValChannel:   "信道: 可选 auto 或 1-233",
	ValSSIDEmpty: "SSID 不能为空",
	ValSSIDLong:  "SSID 超过 32 字节",
	ValPassword:  "密码长度需为 8-63 个字符",
	Hint24:       "auto, 1-13",
	Hint5:        "auto, 36-165",
	Hint6:        "auto, 1-233 步长 4",
	Hint60:       "auto, 1-6",
	HintOther:    "auto",
	Open:         "开放",

	CronTabList: "任务",
	CronTabNew:  "新增",
	CronTabEdit: "编辑",
	CronNone:    "暂无任务",
	CronKind:    "执行方式",
	CronEvery:   "间隔",
	CronAt:      "时间",
	CronExpr:    "表达式",
	CronWeekday: "星期",
	CronCommand: "命令",
	CronAdded:   "已添加任务",
	CronUpdated: "任务已更新",
	CronNoCron:  "本机没有 crontab",
	CronKinds:   [5]string{"每 N 分钟", "每 N 小时", "每天", "每周", "自定义"},
	CronDays:    [7]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"},

	StartupNoTasks: "暂无任务",
	StartupSaved:   "已保存，重启后生效",
	StartupMiss:    "无法读取 %s",

	ValCronRange: "间隔超出范围",
	ValCronTime:  "时间格式应为 HH:MM",
	ValCronExpr:  "表达式需要 5 个字段",

	SysHint:         "危险操作，需要二次确认",
	SysReboot:       "重启",
	SysRebootDesc:   "重启设备",
	SysPoweroff:     "关机",
	SysPoweroffDesc: "断开电源",
	SysFactory:      "恢复出厂设置",
	SysFactoryDesc:  "清空 overlay 后重启",
	SysConfirm:      "确认%s？ enter 执行，esc 取消",
	SysRunning:      "执行中…",
	SysDone:         "命令已提交",
	ValCommand:      "命令不能为空",
}

// Tr is the active translation. Only SetLang/SetLangChoice write it, both
// before a frame is drawn or in response to an explicit user choice.
var Tr = en

// langPick is what the user asked for (auto/en/zh), langAuto is what auto
// resolved to at startup. They are separate so switching to "auto" in the
// settings tab never has to hit the network again.
var (
	langPick     = "auto"
	langAutoPick = "en"
)

var langChoices = []string{"auto", "en", "zh"}

// SetLang is the startup entry point: "auto" asks an IP geolocation service
// which country the public address belongs to, falling back to the locale
// variables when the lookup fails or the router is offline.
func SetLang(name string) {
	langPick = normalizeLang(name)
	if langPick == "auto" {
		langAutoPick = autoLangCode()
	}
	Tr = stringsFor(resolveLang())
}

// SetLangChoice switches language at runtime without any network access.
func SetLangChoice(name string) {
	langPick = normalizeLang(name)
	Tr = stringsFor(resolveLang())
}

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
func LangCode() string {
	return resolveLang()
}

// LangAutoCode is what "auto" resolved to at startup.
func LangAutoCode() string { return langAutoPick }

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
