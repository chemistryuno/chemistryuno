package handlers

import "testing"

func TestIsBuiltinDeckCardSymbolCaseInsensitive(t *testing.T) {
	cases := []string{"Au", "au", "AU", " Na ", "+2"}
	for _, c := range cases {
		if !isBuiltinDeckCardSymbol(c) {
			t.Fatalf("expected builtin symbol for %q", c)
		}
	}
}

func TestIsBuiltinDeckCardSymbolRejectsPluginLikeSymbols(t *testing.T) {
	cases := []string{"", "PLUGIN_X", "Xx", "custom"}
	for _, c := range cases {
		if isBuiltinDeckCardSymbol(c) {
			t.Fatalf("did not expect builtin symbol for %q", c)
		}
	}
}
