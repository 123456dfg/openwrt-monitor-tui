package zh

import "openwrt-monitor-tui/ui/i18n/table"

// T is the Chinese string table.
var T = table.Strings{
	Tabs: [4]string{"状态", "网络", "进程", "设置"},

	Loading:         "加载中…",
	HintScroll:      "tab 切页  ↓/j 滚动  r 刷新  q 退出",
	HintProcs:       "M 按内存排序  P 按CPU排序  ↓/j 滚动  r 刷新  q 退出",
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

	NetTabOverview: "概览", NetTabSpeed: "测速", NetTabPorts: "端口探测",
	HintSpeed: "enter 开始测速  esc 返回  q 退出",
	HintPorts: "enter 开始扫描  esc 返回  q 退出",

	SpeedStart: "开始测速", SpeedRunning: "测速中…", SpeedTool: "工具",
	SpeedServer: "服务器", SpeedDown: "下行", SpeedUp: "上行", SpeedErr: "测速失败",

	PortStart: "开始扫描", PortRunning: "扫描中…", PortProbed: "已探测",
	PortSummary: "发现 %d 台主机，%d 个开放端口", PortNone: "未发现开放端口",

	PID: "进程号", CPUPct: "CPU%", RSS: "内存", Command: "命令",
	NoProcs:     "未找到进程",
	ProcSummary: "共 %d 个进程，显示前 %d",

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
