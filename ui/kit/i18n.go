package kit

import "openwrt-monitor-tui/ui/kit/i18n"

// Tr is the active translation. It is the same *Strings the i18n package
// mutates in place, so a language switch is visible here without re-exports.
var Tr = i18n.Tr

// The language entry points live in ui/kit/i18n together with the string
// tables; these wrappers keep the historical kit.SetLang call sites.
func SetLang(name string)    { i18n.SetLang(name) }
func SetLangChoice(s string) { i18n.SetLangChoice(s) }
func LangChoice() int        { return i18n.LangChoice() }
func LangCode() string       { return i18n.LangCode() }
func LangAutoCode() string   { return i18n.LangAutoCode() }
