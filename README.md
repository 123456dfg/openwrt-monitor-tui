# openwrt-monitor-tui

面向 OpenWrt 路由器的终端状态面板。单个静态二进制，无 CGO、无 libc 依赖，SSH 或串口登录后直接运行；交叉编译覆盖 MIPS（大小端）和 x86，适用于 OpenWrt 19.07 及以上。

已在 **x86** 和 **ZBT-WG3526（MT7621，16M flash）** 上测试可用。

## 主要功能

- **Status**：主机名、型号、内核、在线时间、负载，CPU 占用与频率、内存、交换（未配置时不显示）、温度、磁盘用量；CPU/内存/交换各带一条历史曲线。
- **Processes**：进程列表，按内存或 CPU% 排序（默认降序，表头可点击），显示前 12 个。
- **Network**：接口状态、IP、实时收/发速率、累计流量，每个接口带一条流量曲线；测速子页签基于 speedtest-go 实现 Speedtest.net 协议，实时显示上下行，无需外部工具。
- **Settings**：左侧子页签，键盘和鼠标都能操作。
  - **Wi-Fi**：开关、SSID、密码、信道、加密方式，按 2.4G/5G 分页；Apply 写入 UCI 并 reload。
  - **定时任务**：列出、新增、修改任务，支持启停和“每 N 分钟 / 每 N 小时 / 每天 / 每周 / 自定义”触发。
  - **开机任务**：只列 `/etc/rc.local` 中的任务行，支持启停、编辑、增删。
  - **系统**：重启、关机、恢复出厂设置，均需二次确认。
  - **语言**：Auto / English / 中文，`--lang auto` 按公网 IP 归属地判断。
- 低配适配：低于 1GHz 自动关颜色、放宽采集节奏；按页签懒采样；crontab 和 `rc.local` 后台解析并缓存到 `/tmp`。
- 中英文界面，CJK 列宽对齐；关闭颜色时选中项仍有明确标记。

## 编译

需要 Go 1.24 以上。

```sh
make all          # dist/openwrt-monitor-tui-{mipsle,mips,386,amd64}
make pack         # UPX 压缩 dist 下的产物（约省 60% 体积，需安装 upx）
make release      # all + pack，CI/tag 发布用
make mipsle       # 只编一个目标
make host         # 本机调试
make vet test
make all VERSION=v1.2.3   # 版本号写入 --version
```

| 产物 | 适用设备 |
| --- | --- |
| `openwrt-monitor-tui-mipsle` | ramips/mt7621、mt76x8（K2P、Newifi3、小米路由等，绝大多数家用路由） |
| `openwrt-monitor-tui-mips` | ath79/ar71xx 等大端设备 |
| `openwrt-monitor-tui-386` | OpenWrt 19.07 的 x86 镜像（默认 32 位） |
| `openwrt-monitor-tui-amd64` | x86_64 镜像 / 本机 |

不确定用哪个：`grep DISTRIB_ARCH /etc/openwrt_release`（`mipsel_24kc` → mipsle，`mips_24kc` → mips，`i386` → 386）。

运行：

```sh
scp dist/openwrt-monitor-tui-mipsle root@192.168.1.1:/tmp/
ssh root@192.168.1.1 'chmod +x /tmp/openwrt-monitor-tui-mipsle && /tmp/openwrt-monitor-tui-mipsle'
```

常用参数：`--refresh 1s`、`--lang auto|en|zh`、`--color auto|on|off`、`--ascii`、`--mouse=false`、`--version`。

## 已知不足

- 仅面向 OpenWrt，依赖 `/proc`、`/sys` 和 `uci`/`wifi`/`cron` 脚本；没有接 ubus，看不到无线客户端和服务状态。
- Wi-Fi 只编辑每个 radio 的第一个 AP；信道和加密能否生效取决于驱动与区域设置，错误要等 `wifi reload` 才报。
- 定时任务只能新增、修改和停用，不能删除；开机任务只处理 `/etc/rc.local`，不涉及 `/etc/init.d`。
- 重启、关机、恢复出厂需要 root，执行后连接会断开；恢复出厂会清空全部配置。
- 老 MIPS 目标（如 19.07 的 ramips）内核常未编入 cpufreq，读不到实时频率；温度节点也因设备而异。
- 进程列表与曲线按 1–3 秒采样，数值是区间平均；曲线只保存在内存中，退出即清空。
