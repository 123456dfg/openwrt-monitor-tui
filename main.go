package main

import (
	"flag"
	"fmt"
	"openwrt-monitor-tui/ui/i18n"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"openwrt-monitor-tui/collect"
	_ "openwrt-monitor-tui/prep"
	"openwrt-monitor-tui/ui"
	"openwrt-monitor-tui/ui/style"
)

// version is injected with -ldflags "-X main.version=..." by the release build.
var version = "dev"

func main() {
	started := time.Now()
	refresh := flag.Duration("refresh", time.Second, "refresh interval")
	showVersion := flag.Bool("version", false, "print the version and exit")
	ascii := flag.Bool("ascii", false, "use ASCII glyphs (serial console / non-UTF-8 terminals)")
	mouse := flag.Bool("mouse", true, "mouse support: click tabs, wheel scroll (-mouse=false to keep text selection)")
	lang := flag.String("lang", "auto", "interface language: auto, en, zh")
	color := flag.String("color", "auto", "colors: auto, on, off (auto keeps them off below 1 GHz)")
	flag.Parse()

	if *showVersion {
		fmt.Println("openwrt-monitor-tui " + version)
		return
	}

	i18n.SetLang(*lang)
	if *ascii {
		style.SetASCII()
	}

	switch strings.ToLower(*color) {
	case "on", "always", "yes":
		style.SetColor(true)
	case "off", "never", "no":
		style.SetColor(false)
	default:
		// NO_COLOR is the de-facto standard opt-out; honour it in auto mode.
		style.SetColor(!collect.LowEnd() && os.Getenv("NO_COLOR") == "")
	}

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	if *mouse {
		opts = append(opts, tea.WithMouseCellMotion())
	}

	p := tea.NewProgram(ui.New(*refresh, started), opts...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "openwrt-monitor-tui:", err)
		os.Exit(1)
	}
}
