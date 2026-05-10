package deckcfg

import "testing"

func TestBuiltinDeckDefaultsReturnsCopy(t *testing.T) {
	defaults := BuiltinDeckDefaults()
	defaults["H"] = 999

	again := BuiltinDeckDefaults()
	if again["H"] == 999 {
		t.Fatalf("BuiltinDeckDefaults should return a copy, got shared map")
	}
}

func TestNormalizeBuiltinDeckCardsFiltersUnknownAndInvalidCount(t *testing.T) {
	in := map[string]int{
		"H":   12,
		"+2":  8,
		"Au":  4,
		"XYZ": 3,
		"O":   0,
		"N":   -1,
	}

	normalized, unknown := NormalizeBuiltinDeckCards(in)
	if len(unknown) != 1 || unknown[0] != "XYZ" {
		t.Fatalf("unexpected unknown cards: %#v", unknown)
	}

	if normalized["H"] != 12 || normalized["+2"] != 8 || normalized["Au"] != 4 {
		t.Fatalf("unexpected normalized cards: %#v", normalized)
	}
	if _, ok := normalized["O"]; ok {
		t.Fatalf("card with zero count should be removed")
	}
	if _, ok := normalized["N"]; ok {
		t.Fatalf("card with negative count should be removed")
	}
}

func TestBuiltinDeckDefaultsMatchesTargetScope(t *testing.T) {
	defaults := BuiltinDeckDefaults()
	if _, ok := defaults["reverse"]; ok {
		t.Fatalf("reverse should not be part of builtin deck defaults")
	}
	if _, ok := defaults["skip"]; ok {
		t.Fatalf("skip should not be part of builtin deck defaults")
	}
	for _, symbol := range []string{"Au", "+2", "+4", "He", "Ne", "Ar", "Kr"} {
		if _, ok := defaults[symbol]; !ok {
			t.Fatalf("missing builtin special card: %s", symbol)
		}
	}
}

func TestIsLegacyOneCountGlobalDeck(t *testing.T) {
	legacy := map[string]int{
		"H": 1, "O": 1,
		"C": 1, "N": 1, "F": 1, "Na": 1, "Mg": 1, "Al": 1,
		"Si": 1, "P": 1, "S": 1, "Cl": 1, "K": 1, "Ca": 1,
		"Mn": 1, "Fe": 1, "Cu": 1, "Zn": 1, "Br": 1, "I": 1, "Ag": 1,
	}
	if !IsLegacyOneCountGlobalDeck(legacy) {
		t.Fatalf("expected legacy deck to be detected")
	}

	withSpecial := BuiltinDeckDefaults()
	if IsLegacyOneCountGlobalDeck(withSpecial) {
		t.Fatalf("builtin default deck should not be treated as legacy")
	}
}
