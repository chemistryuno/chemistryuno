package utils

import (
	"regexp"
	"testing"
)

func TestGenerateRandomNickname(t *testing.T) {
	defaultNickname := GenerateRandomNickname("")
	if !regexp.MustCompile(`^研究员\d{6}$`).MatchString(defaultNickname) {
		t.Fatalf("default nickname has unexpected format: %q", defaultNickname)
	}

	customNickname := GenerateRandomNickname("化学家")
	if !regexp.MustCompile(`^化学家\d{6}$`).MatchString(customNickname) {
		t.Fatalf("custom nickname has unexpected format: %q", customNickname)
	}
}
