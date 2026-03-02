package deckcfg

import (
	"sort"
	"strings"
)

var builtinNormalDeckDefaults = map[string]int{
	"H": 12,
	"O": 12,

	"C": 4, "N": 4, "F": 4, "Na": 4, "Mg": 4, "Al": 4,
	"Si": 4, "P": 4, "S": 4, "Cl": 4, "K": 4, "Ca": 4,
	"Mn": 4, "Fe": 4, "Cu": 4, "Zn": 4, "Br": 4, "I": 4, "Ag": 4,
}

var builtinSpecialDeckDefaults = map[string]int{
	"+2": 8,
	"+4": 4,
	"Au": 4,
	"He": 1,
	"Ne": 1,
	"Ar": 1,
	"Kr": 1,
}

var builtinDeckDefaults = func() map[string]int {
	result := make(map[string]int, len(builtinNormalDeckDefaults)+len(builtinSpecialDeckDefaults))
	for key, value := range builtinNormalDeckDefaults {
		result[key] = value
	}
	for key, value := range builtinSpecialDeckDefaults {
		result[key] = value
	}
	return result
}()

var builtinDeckCardSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(builtinDeckDefaults))
	for key := range builtinDeckDefaults {
		m[key] = struct{}{}
	}
	return m
}()

// BuiltinDeckDefaults returns a copy of default built-in deck cards.
func BuiltinDeckDefaults() map[string]int {
	result := make(map[string]int, len(builtinDeckDefaults))
	for key, value := range builtinDeckDefaults {
		result[key] = value
	}
	return result
}

// IsLegacyOneCountGlobalDeck returns true for the old broken global default:
// all normal cards are 1 and all special cards are missing.
func IsLegacyOneCountGlobalDeck(cards map[string]int) bool {
	if len(cards) != len(builtinNormalDeckDefaults) {
		return false
	}

	for symbol, count := range cards {
		if _, isNormal := builtinNormalDeckDefaults[symbol]; !isNormal {
			return false
		}
		if count != 1 {
			return false
		}
	}

	for symbol := range builtinNormalDeckDefaults {
		if cards[symbol] != 1 {
			return false
		}
	}

	return true
}

// NormalizeBuiltinDeckCards keeps only built-in cards and positive counts.
// It also returns unknown card symbols that were filtered out.
func NormalizeBuiltinDeckCards(cards map[string]int) (map[string]int, []string) {
	normalized := make(map[string]int)
	if len(cards) == 0 {
		return normalized, []string{}
	}

	unknown := make([]string, 0)
	seenUnknown := make(map[string]struct{})
	for rawKey, count := range cards {
		key := strings.TrimSpace(rawKey)
		if key == "" {
			continue
		}
		if _, ok := builtinDeckCardSet[key]; !ok {
			if _, exists := seenUnknown[key]; !exists {
				seenUnknown[key] = struct{}{}
				unknown = append(unknown, key)
			}
			continue
		}
		if count <= 0 {
			continue
		}
		normalized[key] = count
	}
	sort.Strings(unknown)
	return normalized, unknown
}
