package game

import (
	"reflect"
	"testing"

	"chemistryuno/backend/models"
)

func TestNormalizeSubscripts(t *testing.T) {
	got := NormalizeSubscripts("Ca(OH)₂ + H₂SO₄")
	want := "Ca(OH)2 + H2SO4"
	if got != want {
		t.Fatalf("NormalizeSubscripts() = %q, want %q", got, want)
	}
}

func TestParseSubstanceHandlesCountsAndGroups(t *testing.T) {
	tests := []struct {
		formula string
		want    map[string]int
	}{
		{"H2O", map[string]int{"H": 2, "O": 1}},
		{"Ca(OH)2", map[string]int{"Ca": 1, "O": 2, "H": 2}},
		{"Al2(SO4)3", map[string]int{"Al": 2, "S": 3, "O": 12}},
		{"H₂SO₄", map[string]int{"H": 2, "S": 1, "O": 4}},
	}

	for _, tt := range tests {
		t.Run(tt.formula, func(t *testing.T) {
			if got := parseSubstance(tt.formula); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseSubstance() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCanFormSubstanceRequiresElementKindsOnly(t *testing.T) {
	elements := map[string]int{"H": 1, "O": 1}
	if !canFormSubstance("H2O", elements) {
		t.Fatal("expected H and O cards to form H2O even without exact coefficients")
	}
	if canFormSubstance("HCl", elements) {
		t.Fatal("expected missing Cl to reject HCl")
	}
}

func TestCanReactSpecialSubstances(t *testing.T) {
	tests := [][2]string{
		{"He", "NaCl"},
		{"NaCl", "+2"},
		{"reverse", "H2O"},
		{"skip", "CuSO4"},
	}
	for _, tt := range tests {
		if !CanReact(tt[0], tt[1]) {
			t.Fatalf("expected %s to react with %s as a special substance", tt[0], tt[1])
		}
	}
	if CanReact("NaCl", "H2O") {
		t.Fatal("expected ordinary substances without database reaction to be rejected")
	}
}

func TestGetSubstancesFromElementsIncludesFunctionalCardsAndApprovedSubstances(t *testing.T) {
	db := setupSubstanceValidationTest(t)
	seedApprovedSubstance(t, db, "H")
	RebuildSubstanceCache()

	got := GetSubstancesFromElements([]models.Card{
		{Type: "H"},
		{Type: "O"},
		{Type: "Ar"},
		{Type: "Au"},
	})
	set := map[string]bool{}
	for _, value := range got {
		set[value] = true
	}

	for _, want := range []string{"H", "Ar", "Au"} {
		if !set[want] {
			t.Fatalf("expected %q in %#v", want, got)
		}
	}
	if set["O"] || set["O2"] {
		t.Fatalf("ordinary O substances should require database approval, got %#v", got)
	}
}

func TestJudgeReactionRepresentativeRules(t *testing.T) {
	tests := []struct {
		name string
		s1   string
		s2   string
		want bool
	}{
		{"acid base neutralization", "HCl", "NaOH", true},
		{"metal acid", "Zn", "HCl", true},
		{"water active metal", "H2O", "Na", true},
		{"inert gas", "Ar", "HCl", false},
		{"stable nonmetal oxygen", "N2", "O2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JudgeReaction(tt.s1, tt.s2); got != tt.want {
				t.Fatalf("JudgeReaction(%q, %q) = %v, want %v", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}
