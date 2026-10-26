# openwrt-monitor-tui

[简体中文](docs/zh-cn/README.md)

A terminal status panel for OpenWrt routers. Ships as a single static binary — no CGO, no libc dependency — ready to run over SSH or serial; cross-compilation covers MIPS (both endiannesses) and x86, and targets OpenWrt 19.07 and newer.

Tested on **x86** and **ZBT-WG3526 (MT7621, 16M flash)**.

## Features

- **Status**: hostname, model, kernel, uptime, load, CPU usage and frequency, memory, swap (hidden when unconfigured), temperature, disk usage; a history chart each for CPU, memory and swap.
- **Processes**: process list sorted by memory or CPU% (descending by default, clickable header), top 12.
- **Network**: interface state, addresses, live rx/tx rates, traffic totals, and a traffic chart per interface. Two sub-tabs beside the overview:
  - **Speed test**: the Speedtest.net protocol implemented in-process (speedtest-go) — download and upload with live rates and the server region, no external tools required.
  - **Ports**: discovers downstream LAN hosts from the kernel ARP table and TCP-probes common service ports on each, as a background job with live progress.
- **Settings**: sidebar sub-tabs, keyboard and mouse.
  - **Wi-Fi**: switch, SSID, password, channel, security, paged by 2.4G/5G; Apply writes UCI and reloads.
  - **Scheduled**: list, add and edit cron jobs, with enable/disable and interval / daily / weekly / custom triggers.
  - **Startup**: lists the task lines of `/etc/rc.local`, with toggles and editing.
  - **System**: reboot, power off, factory reset — all double-confirmed.
  - **Language**: Auto / English / 中文; `--lang auto` picks by the public IP's country.
- Low-end friendly: colours off and relaxed sampling below 1 GHz; lazy per-tab sampling; crontab and `rc.local` parsed in the background and cached in `/tmp`.
- English & Chinese interface, CJK-aware column widths; clear selection markers even with colours off.

## Build

Requires Go 1.24+.

```sh
make all          # dist/openwrt-monitor-tui-{mipsle,mips,386,amd64}
make pack         # UPX-compress the dist artifacts (~60% smaller, needs upx)
make release      # all + pack, used by CI / tag releases
make mipsle       # a single target
make host         # local debugging
make vet test
make all VERSION=v1.2.3   # stamp the version into --version
```

| artifact | target |
| --- | --- |
| `openwrt-monitor-tui-mipsle` | ramips/mt7621, mt76x8 (K2P, Newifi3, Xiaomi routers — most home devices) |
| `openwrt-monitor-tui-mips` | ath79/ar71xx big-endian |
| `openwrt-monitor-tui-386` | OpenWrt 19.07 x86 images (32-bit default) |
| `openwrt-monitor-tui-amd64` | x86_64 images / local |

Not sure which: `grep DISTRIB_ARCH /etc/openwrt_release` (`mipsel_24kc` → mipsle, `mips_24kc` → mips, `i386` → 386).

Run:

```sh
scp dist/openwrt-monitor-tui-mipsle root@192.168.1.1:/tmp/
ssh root@192.168.1.1 'chmod +x /tmp/openwrt-monitor-tui-mipsle && /tmp/openwrt-monitor-tui-mipsle'
```

Common flags: `--refresh 1s`, `--lang auto|en|zh`, `--color auto|on|off`, `--ascii`, `--mouse=false`, `--version`.

## Known limitations

- OpenWrt only: relies on `/proc`, `/sys` and the `uci`/`wifi`/`cron` scripts; no ubus, so no wireless client or service status.
- Wi-Fi edits only the first AP of each radio; whether channel/security changes take effect depends on driver and region, and errors surface only after `wifi reload`.
- Scheduled jobs can be added, edited and disabled but not deleted; startup covers `/etc/rc.local` only, not `/etc/init.d`.
- Reboot / power off / factory reset need root and drop the connection; factory reset wipes all configuration.
- Old MIPS targets (e.g. 19.07 ramips) often lack cpufreq, so live frequency is unavailable; thermal nodes vary by device.
- Process list and charts sample every 1–3 s and show interval averages; charts live in memory only and clear on exit.

## Disclaimer

The speed test and port scan features actively probe networks. Only use them on networks you own or are explicitly authorized to test — **users are responsible for ensuring compliance with local laws and regulations** when using these tools.
