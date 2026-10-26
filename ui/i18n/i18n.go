// Package i18n owns the language selection and the shared translation
// pointer. The string tables themselves live in the per-language subpackages
// (en/, zh/); the schema they fill in is ui/i18n/table.
package i18n

import (
	"strings"

	i18nen "openwrt-monitor-tui/ui/i18n/en"
	"openwrt-monitor-tui/ui/i18n/table"
	i18nzh "openwrt-monitor-tui/ui/i18n/zh"
)

// Strings is the schema alias, kept so the rest of the tree can spell the
// translation type without reaching into the table package.
type Strings = table.Strings

// Tr is the active translation. It is a pointer that kit shares: the setters
// below replace the pointed-to value rather than the pointer, so every reader
// sees the new language without any re-export step.
var Tr = new(table.Strings)

func init() { *Tr = i18nen.T }

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

func stringsFor(code string) table.Strings {
	if code == "zh" {
		return i18nzh.T
	}
	return i18nen.T
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
