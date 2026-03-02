package game

import "chemistryuno/backend/deckcfg"

// BuiltinDeckDefaults returns a copy of default built-in deck cards.
func BuiltinDeckDefaults() map[string]int {
	return deckcfg.BuiltinDeckDefaults()
}

// NormalizeBuiltinDeckCards keeps only built-in cards and positive counts.
// It also returns unknown card symbols that were filtered out.
func NormalizeBuiltinDeckCards(cards map[string]int) (map[string]int, []string) {
	return deckcfg.NormalizeBuiltinDeckCards(cards)
}
