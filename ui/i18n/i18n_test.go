package i18n

import "testing"

// SetLangChoice must swap the shared translation in place, so a re-export like
// style.Tr keeps pointing at the current language.
func TestSwitchInPlace(t *testing.T) {
	t.Cleanup(func() { SetLangChoice("en") })
	before := Tr
	SetLangChoice("zh")
	if Tr != before {
		t.Fatal("Tr pointer was replaced instead of mutated")
	}
	if Tr.Tabs[0] != "状态" {
		t.Fatalf("zh not applied: %q", Tr.Tabs[0])
	}
	SetLangChoice("en")
	if Tr.Tabs[0] != "Status" {
		t.Fatalf("en not applied: %q", Tr.Tabs[0])
	}
}
